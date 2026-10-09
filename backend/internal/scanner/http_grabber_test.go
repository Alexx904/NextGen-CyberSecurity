package scanner

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestGrabHTTPInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "TP-LINK Web Server")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head><title>TP-Link Wireless Router</title></head><body>Welcome</body></html>"))
	}))
	defer ts.Close()

	// Estrai host e porta
	host, portStr, err := net.SplitHostPort(ts.Listener.Addr().String())
	if err != nil {
		t.Fatalf("errore parse hostport: %v", err)
	}
	port, _ := strconv.Atoi(portStr)

	title, server := GrabHTTPInfo(context.Background(), host, []int{port})
	if title != "TP-Link Wireless Router" {
		t.Errorf("title = %q; atteso 'TP-Link Wireless Router'", title)
	}
	if server != "TP-LINK Web Server" {
		t.Errorf("server = %q; atteso 'TP-LINK Web Server'", server)
	}
}

func TestGrabHTTPInfo_NoWebPorts(t *testing.T) {
	title, server := GrabHTTPInfo(context.Background(), "127.0.0.1", []int{22, 53})
	if title != "" || server != "" {
		t.Errorf("attese stringhe vuote per porte non web, ottenuto title=%q, server=%q", title, server)
	}
}

func TestGrabHTTPInfo_UnreachableHost(t *testing.T) {
	title, server := GrabHTTPInfo(context.Background(), "192.0.2.1", []int{80})
	if title != "" || server != "" {
		t.Errorf("attese stringhe vuote per host non raggiungibile, ottenuto title=%q, server=%q", title, server)
	}
}

