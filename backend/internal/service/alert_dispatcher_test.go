package service

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"nextgen-cybersec/internal/domain"
)

type alertRoundTripper func(*http.Request) (*http.Response, error)

func (f alertRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type alertTestResolver struct{ ips []netip.Addr }

func (r alertTestResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.ips, nil
}

func TestAlertDestinationBoundaries(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.1.1.1", "172.16.1.1", "192.168.1.1", "169.254.169.254", "100.64.1.1", "0.0.0.0", "192.0.2.1", "198.18.1.1", "240.0.0.1", "::1", "::ffff:127.0.0.1", "fe80::1", "fd00::1", "64:ff9b::a00:1", "2001:db8::1", "2002:a00:1::"} {
		if publicAlertAddress(netip.MustParseAddr(ip)) {
			t.Errorf("unsafe address accepted: %s", ip)
		}
	}
	for _, ip := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !publicAlertAddress(netip.MustParseAddr(ip)) {
			t.Errorf("public address refused: %s", ip)
		}
	}
	for _, answers := range [][]netip.Addr{nil, {netip.MustParseAddr("127.0.0.1")}, {netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}} {
		if conn, err := alertDialer(alertTestResolver{answers})(context.Background(), "tcp", "rebind.example:443"); err == nil || conn != nil {
			t.Fatal("unsafe DNS answer dialed")
		}
	}
	secret := strings.Repeat("test-key", 5)
	for _, endpoint := range []string{"http://example.com", "https://user:pass@example.com", "https://example.com:8443", "https://127.0.0.1/", "https://[::ffff:127.0.0.1]/", "https://example.com/#fragment"} {
		if ValidateAlertConfig(domain.AlertWebhook, domain.AlertChannelConfig{Endpoint: endpoint, HMACSecret: secret}) == nil {
			t.Error("unsafe endpoint accepted", endpoint)
		}
	}
}

func TestAlertHMACExactBytes(t *testing.T) {
	got := AlertSignature("0123456789abcdef0123456789abcdef", "1700000000", "11111111-1111-4111-8111-111111111111", []byte(`{"test":true}`))
	// Independent HMACSHA256 vector generated with .NET, including separators.
	if got != "sha256=c1fb82dd708b5740a7ade7b0ead890d2269ff10706a3ca2b8b83dee2151d44a2" {
		t.Fatal(got)
	}
	if got == AlertSignature("0123456789abcdef0123456789abcdef", "1700000001", "11111111-1111-4111-8111-111111111111", []byte(`{"test":true}`)) {
		t.Fatal("timestamp not signed")
	}
}

func TestAlertProviderHandling(t *testing.T) {
	event := domain.AlertEvent{ID: uuid.New(), TenantID: uuid.New(), Kind: "channel.test", OccurredAt: time.Now()}
	for _, tc := range []struct {
		name       string
		kind       domain.AlertChannelKind
		status     int
		body, code string
		retry      bool
	}{
		{"slack ok", domain.AlertSlack, 200, "ok", "ACCEPTED", false},
		{"slack rejection", domain.AlertSlack, 200, "private-secret-body", "INVALID_RESPONSE", false},
		{"teams acceptance", domain.AlertTeams, 202, "", "ACCEPTED", false},
		{"telegram ok", domain.AlertTelegram, 200, `{"ok":true}`, "ACCEPTED", false},
		{"telegram failed body", domain.AlertTelegram, 200, `{"ok":false,"error_code":400}`, "REMOTE_REJECTED", false},
		{"generic retry", domain.AlertWebhook, 503, "private-secret-body", "REMOTE_RETRY", true},
		{"generic forbidden", domain.AlertWebhook, 403, "private-secret-body", "REMOTE_REJECTED", false},
		{"redirect refused", domain.AlertWebhook, 302, "", "REMOTE_REJECTED", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := newAlertHTTPClient()
			client.Transport = alertRoundTripper(func(r *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(r.Body)
				if r.Header.Get("Content-Type") != "application/json" || len(body) == 0 {
					t.Fatal("missing JSON")
				}
				if tc.kind == domain.AlertWebhook && (r.Header.Get("X-NextGen-Event-ID") != event.ID.String() || !strings.HasPrefix(r.Header.Get("X-NextGen-Signature"), "sha256=")) {
					t.Fatal("missing event signature")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: http.Header{"Location": []string{"https://127.0.0.1/private"}}, Request: r}, nil
			})
			d := &alertDispatcher{client: client, now: time.Now}
			config := domain.AlertChannelConfig{Endpoint: "https://hooks.slack.com/test-only", HMACSecret: strings.Repeat("synthetic", 5), Token: "123456:" + strings.Repeat("a", 24), ChatID: "123456"}
			result := d.Send(context.Background(), tc.kind, config, event)
			if result.Code != tc.code || result.Retry != tc.retry || calls != 1 {
				t.Fatalf("result %+v calls %d", result, calls)
			}
		})
	}
}

func TestAlertSMTPRefusesPlaintextAndCancels(t *testing.T) {
	config := domain.AlertChannelConfig{SMTPHost: "smtp.example.test", SMTPPort: 587, SMTPFrom: "from@example.test", SMTPTo: []string{"to@example.test"}, SMTPUsername: "test", SMTPPassword: "test-only"}
	d := &alertDispatcher{dial: func(context.Context, string, string) (net.Conn, error) {
		client, server := net.Pipe()
		go func() {
			defer server.Close()
			server.SetDeadline(time.Now().Add(time.Second))
			io.WriteString(server, "220 smtp fixture\r\n")
			line, _ := bufio.NewReader(server).ReadString('\n')
			if !strings.HasPrefix(line, "EHLO ") {
				return
			}
			io.WriteString(server, "250-fixture\r\n250 HELP\r\n")
		}()
		return client, nil
	}}
	if result := d.Send(context.Background(), domain.AlertSMTP, config, domain.AlertEvent{}); result.Code != "SMTP_TLS_REQUIRED" {
		t.Fatalf("plaintext relay accepted: %+v", result)
	}
	d.dial = func(context.Context, string, string) (net.Conn, error) {
		client, server := net.Pipe()
		t.Cleanup(func() { server.Close() })
		return client, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	result := d.Send(ctx, domain.AlertSMTP, config, domain.AlertEvent{})
	if result.Code != "SMTP_FAILED" || time.Since(start) > time.Second {
		t.Fatal("cancellation did not interrupt SMTP")
	}
}

func TestAlertTelegramRateLimitDelay(t *testing.T) {
	d := &alertDispatcher{now: time.Now, client: &http.Client{Transport: alertRoundTripper(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"ok":false,"error_code":429,"parameters":{"retry_after":90}}`)), Request: r}, nil
	})}}
	config := domain.AlertChannelConfig{Token: "123456:" + strings.Repeat("a", 24), ChatID: "123456"}
	result := d.Send(context.Background(), domain.AlertTelegram, config, domain.AlertEvent{})
	if !result.Retry || result.RetryAfter != 90*time.Second {
		t.Fatalf("provider delay lost: %+v", result)
	}
}

func TestAlertSMTPRealTLSAcceptance(t *testing.T) {
	certificateSource := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	certificates := certificateSource.TLS.Certificates
	certificate := certificateSource.Certificate()
	certificateSource.Close()
	roots := x509.NewCertPool()
	roots.AddCert(certificate)
	for _, port := range []int{465, 587} {
		t.Run(strconv.Itoa(port), func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			type outcome struct {
				body string
				err  error
			}
			done := make(chan outcome, 1)
			go func() {
				conn, err := listener.Accept()
				if err != nil {
					done <- outcome{err: err}
					return
				}
				defer conn.Close()
				conn.SetDeadline(time.Now().Add(3 * time.Second))
				secure := port == 465
				if secure {
					conn = tls.Server(conn, &tls.Config{Certificates: certificates, MinVersion: tls.VersionTLS12})
				}
				io.WriteString(conn, "220 local SMTP fixture\r\n")
				reader := bufio.NewReader(conn)
				body := ""
				dataMode := false
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						done <- outcome{err: err}
						return
					}
					if dataMode {
						if line == ".\r\n" {
							dataMode = false
							io.WriteString(conn, "250 accepted\r\n")
						} else {
							body += line
						}
						continue
					}
					switch {
					case strings.HasPrefix(line, "EHLO "):
						if !secure {
							io.WriteString(conn, "250-fixture\r\n250 STARTTLS\r\n")
						} else {
							io.WriteString(conn, "250-fixture\r\n250 AUTH PLAIN\r\n")
						}
					case strings.HasPrefix(line, "STARTTLS"):
						io.WriteString(conn, "220 ready\r\n")
						conn = tls.Server(conn, &tls.Config{Certificates: certificates, MinVersion: tls.VersionTLS12})
						reader = bufio.NewReader(conn)
						secure = true
					case strings.HasPrefix(line, "AUTH "):
						if !secure {
							done <- outcome{err: errors.New("authentication before TLS")}
							return
						}
						io.WriteString(conn, "235 authenticated\r\n")
					case strings.HasPrefix(line, "MAIL FROM:"), strings.HasPrefix(line, "RCPT TO:"):
						io.WriteString(conn, "250 okay\r\n")
					case line == "DATA\r\n":
						dataMode = true
						io.WriteString(conn, "354 send data\r\n")
					case line == "QUIT\r\n":
						io.WriteString(conn, "221 bye\r\n")
						done <- outcome{body: body}
						return
					default:
						done <- outcome{err: errors.New("unexpected SMTP command")}
						return
					}
				}
			}()
			d := &alertDispatcher{tlsRoots: roots, dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "tcp", listener.Addr().String())
			}}
			config := domain.AlertChannelConfig{SMTPHost: "example.com", SMTPPort: port, SMTPFrom: "from@example.test", SMTPTo: []string{"to@example.test"}, SMTPUsername: "fixture", SMTPPassword: "test-only"}
			event := domain.AlertEvent{ID: uuid.New(), Kind: "channel.test", EntityID: "synthetic-only"}
			result := d.Send(context.Background(), domain.AlertSMTP, config, event)
			if result.Code != "ACCEPTED" {
				t.Fatalf("SMTP TLS failed: %+v", result)
			}
			select {
			case actual := <-done:
				if actual.err != nil || !strings.Contains(actual.body, event.ID.String()) || strings.Contains(actual.body, config.SMTPPassword) {
					t.Fatalf("SMTP outcome: %v", actual.err)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("SMTP fixture did not finish")
			}
		})
	}
}
