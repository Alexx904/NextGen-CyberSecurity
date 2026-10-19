package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nextgen-cybersec/internal/domain"
)

type AlertSendResult struct {
	Code       string
	Retry      bool
	RetryAfter time.Duration
}
type AlertDispatcher interface {
	Send(context.Context, domain.AlertChannelKind, domain.AlertChannelConfig, domain.AlertEvent) AlertSendResult
}
type alertDispatcher struct {
	client   *http.Client
	dial     func(context.Context, string, string) (net.Conn, error)
	now      func() time.Time
	tlsRoots *x509.CertPool
}

func NewAlertDispatcher() AlertDispatcher {
	return &alertDispatcher{client: newAlertHTTPClient(), dial: alertDialer(net.DefaultResolver), now: time.Now}
}

func ValidateAlertConfig(kind domain.AlertChannelKind, c domain.AlertChannelConfig) error {
	invalid := domain.ErrValidation
	if len(c.Endpoint) > 4096 || len(c.Token) > 256 || len(c.HMACSecret) > 256 || len(c.SMTPPassword) > 1024 || len(c.SMTPUsername) > 254 {
		return invalid
	}
	switch kind {
	case domain.AlertSlack, domain.AlertTeams, domain.AlertWebhook:
		u, err := url.Parse(c.Endpoint)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
			return invalid
		}
		if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicAlertAddress(ip) {
			return invalid
		}
		if kind == domain.AlertSlack && u.Hostname() != "hooks.slack.com" && u.Hostname() != "hooks.slack-gov.com" {
			return invalid
		}
		if kind == domain.AlertWebhook && len(c.HMACSecret) < 32 {
			return invalid
		}
	case domain.AlertTelegram:
		if !regexp.MustCompile(`^[0-9]{5,16}:[A-Za-z0-9_-]{20,120}$`).MatchString(c.Token) || !regexp.MustCompile(`^-?[0-9]{1,20}$|^@[A-Za-z0-9_]{5,64}$`).MatchString(c.ChatID) {
			return invalid
		}
	case domain.AlertSMTP:
		if c.SMTPHost == "" || len(c.SMTPHost) > 253 || strings.ContainsAny(c.SMTPHost, "\r\n/ @:[]") || (c.SMTPPort != 465 && c.SMTPPort != 587) || len(c.SMTPTo) == 0 || len(c.SMTPTo) > 10 {
			return invalid
		}
		if (c.SMTPUsername == "") != (c.SMTPPassword == "") {
			return invalid
		}
		if ip, err := netip.ParseAddr(c.SMTPHost); err == nil && !publicAlertAddress(ip) {
			return invalid
		}
		for _, address := range append([]string{c.SMTPFrom}, c.SMTPTo...) {
			a, err := mail.ParseAddress(address)
			if err != nil || a.Address != address || len(address) > 254 || strings.ContainsAny(address, "\r\n") {
				return invalid
			}
		}
	default:
		return invalid
	}
	return nil
}

// AlertSignature signs exact bytes with a timestamp and stable event ID. Receivers
// should reject stale timestamps and deduplicate event IDs before processing.
func AlertSignature(secret, timestamp, eventID string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + eventID + "."))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func alertMessage(event domain.AlertEvent) string {
	return fmt.Sprintf("NextGen %s | event=%s | entity=%s | host=%s | port=%d | CVE=%s | severity=%s", event.Kind, event.ID, event.EntityID, event.HostIP, event.Port, event.CVEID, event.Severity)
}

func (d *alertDispatcher) Send(ctx context.Context, kind domain.AlertChannelKind, c domain.AlertChannelConfig, event domain.AlertEvent) AlertSendResult {
	if ValidateAlertConfig(kind, c) != nil {
		return AlertSendResult{Code: "INVALID_CONFIGURATION"}
	}
	if kind == domain.AlertSMTP {
		return d.sendSMTP(ctx, c, event)
	}
	endpoint := c.Endpoint
	var payload any
	switch kind {
	case domain.AlertSlack:
		payload = map[string]any{"text": alertMessage(event), "mrkdwn": false}
	case domain.AlertTeams:
		payload = map[string]any{"type": "message", "attachments": []any{map[string]any{"contentType": "application/vnd.microsoft.card.adaptive", "content": map[string]any{"type": "AdaptiveCard", "version": "1.2", "body": []any{map[string]any{"type": "TextBlock", "text": alertMessage(event), "wrap": true}}}}}}
	case domain.AlertTelegram:
		endpoint = "https://api.telegram.org/bot" + c.Token + "/sendMessage"
		payload = map[string]any{"chat_id": c.ChatID, "text": alertMessage(event), "disable_web_page_preview": true}
	case domain.AlertWebhook:
		payload = event
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return AlertSendResult{Code: "INVALID_PAYLOAD"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return AlertSendResult{Code: "INVALID_CONFIGURATION"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NextGen-Alerting/1.0")
	if kind == domain.AlertWebhook {
		timestamp := strconv.FormatInt(d.now().Unix(), 10)
		req.Header.Set("X-NextGen-Timestamp", timestamp)
		req.Header.Set("X-NextGen-Event-ID", event.ID.String())
		req.Header.Set("X-NextGen-Signature", AlertSignature(c.HMACSecret, timestamp, event.ID.String(), body))
	}
	res, err := d.client.Do(req)
	if err != nil {
		return AlertSendResult{Code: "CONNECTION_FAILED", Retry: true}
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil || len(data) > 65536 {
		return AlertSendResult{Code: "INVALID_RESPONSE", Retry: true}
	}
	if res.StatusCode == 429 || res.StatusCode >= 500 {
		delay := time.Duration(0)
		if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 && seconds <= 3600 {
			delay = time.Duration(seconds) * time.Second
		} else if date, err := http.ParseTime(res.Header.Get("Retry-After")); err == nil {
			delay = date.Sub(d.now())
			if delay < 0 {
				delay = 0
			}
			if delay > time.Hour {
				delay = time.Hour
			}
		}
		if kind == domain.AlertTelegram {
			var reply struct {
				Parameters struct {
					RetryAfter int `json:"retry_after"`
				} `json:"parameters"`
			}
			if json.Unmarshal(data, &reply) == nil && reply.Parameters.RetryAfter > 0 {
				telegramDelay := time.Duration(reply.Parameters.RetryAfter) * time.Second
				if telegramDelay > time.Hour {
					telegramDelay = time.Hour
				}
				if telegramDelay > delay {
					delay = telegramDelay
				}
			}
		}
		return AlertSendResult{Code: "REMOTE_RETRY", Retry: true, RetryAfter: delay}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return AlertSendResult{Code: "REMOTE_REJECTED"}
	}
	if kind == domain.AlertSlack && strings.TrimSpace(string(data)) != "ok" {
		return AlertSendResult{Code: "INVALID_RESPONSE"}
	}
	if kind == domain.AlertTelegram {
		var reply struct {
			OK         bool `json:"ok"`
			Parameters struct {
				RetryAfter int `json:"retry_after"`
			} `json:"parameters"`
			ErrorCode int `json:"error_code"`
		}
		if json.Unmarshal(data, &reply) != nil {
			return AlertSendResult{Code: "INVALID_RESPONSE", Retry: true}
		}
		if !reply.OK {
			delay := time.Duration(reply.Parameters.RetryAfter) * time.Second
			if delay < 0 {
				delay = 0
			}
			if delay > time.Hour {
				delay = time.Hour
			}
			return AlertSendResult{Code: "REMOTE_REJECTED", Retry: reply.ErrorCode == 429 || reply.ErrorCode >= 500, RetryAfter: delay}
		}
	}
	return AlertSendResult{Code: "ACCEPTED"}
}

func (d *alertDispatcher) sendSMTP(ctx context.Context, c domain.AlertChannelConfig, event domain.AlertEvent) AlertSendResult {
	fail := func() AlertSendResult { return AlertSendResult{Code: "SMTP_FAILED", Retry: true} }
	conn, err := d.dial(ctx, "tcp", net.JoinHostPort(c.SMTPHost, strconv.Itoa(c.SMTPPort)))
	if err != nil {
		return fail()
	}
	rawConn := conn
	defer rawConn.Close()
	stop := context.AfterFunc(ctx, func() { rawConn.Close() })
	defer stop()
	deadline := time.Now().Add(10 * time.Second)
	if end, ok := ctx.Deadline(); ok && end.Before(deadline) {
		deadline = end
	}
	conn.SetDeadline(deadline)
	tlsConfig := &tls.Config{ServerName: c.SMTPHost, MinVersion: tls.VersionTLS12, RootCAs: d.tlsRoots}
	if c.SMTPPort == 465 {
		secure := tls.Client(conn, tlsConfig)
		if secure.HandshakeContext(ctx) != nil {
			return fail()
		}
		conn = secure
	}
	client, err := smtp.NewClient(conn, c.SMTPHost)
	if err != nil {
		return fail()
	}
	defer client.Close()
	if c.SMTPPort == 587 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return AlertSendResult{Code: "SMTP_TLS_REQUIRED"}
		}
		if client.StartTLS(tlsConfig) != nil {
			return fail()
		}
	}
	if c.SMTPUsername != "" {
		if client.Auth(smtp.PlainAuth("", c.SMTPUsername, c.SMTPPassword, c.SMTPHost)) != nil {
			return AlertSendResult{Code: "SMTP_AUTH_REJECTED"}
		}
	}
	if client.Mail(c.SMTPFrom) != nil {
		return fail()
	}
	for _, recipient := range c.SMTPTo {
		if client.Rcpt(recipient) != nil {
			return fail()
		}
	}
	w, err := client.Data()
	if err != nil {
		return fail()
	}
	message := "From: " + c.SMTPFrom + "\r\nTo: " + strings.Join(c.SMTPTo, ", ") + "\r\nSubject: NextGen security notification\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + alertMessage(event) + "\r\n"
	if _, err = w.Write([]byte(message)); err != nil {
		return fail()
	}
	if w.Close() != nil {
		return fail()
	}
	// DATA success is provider acceptance. QUIT failure must not retry accepted mail.
	_ = client.Quit()
	return AlertSendResult{Code: "ACCEPTED"}
}
