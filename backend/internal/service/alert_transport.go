package service

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"time"
)

var errAlertDestination = errors.New("alert destination rejected")

type alertResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

var alertDeniedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

func publicAlertAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range alertDeniedPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

// Resolve once, validate all answers, then dial a literal IP. TLS still validates
// the original hostname. Redirects and environment proxies are disabled.
func alertDialer(resolver alertResolver) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil || (port != "443" && port != "465" && port != "587") {
			return nil, errAlertDestination
		}
		ips, err := resolver.LookupNetIP(ctx, "ip", host)
		if err != nil || len(ips) == 0 || len(ips) > 32 {
			return nil, errAlertDestination
		}
		for _, ip := range ips {
			if !publicAlertAddress(ip) {
				return nil, errAlertDestination
			}
		}
		dialer := net.Dialer{Timeout: 5 * time.Second}
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			if ctx.Err() != nil {
				break
			}
		}
		return nil, errors.New("alert connection failed")
	}
}

func newAlertHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{Proxy: nil, DialContext: alertDialer(net.DefaultResolver),
			TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
			TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 5 * time.Second,
			MaxIdleConns: 10, MaxConnsPerHost: 2, IdleConnTimeout: 30 * time.Second},
	}
}
