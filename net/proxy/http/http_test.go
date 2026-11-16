package http

import (
	"crypto/tls"
	"net/url"
	"testing"

	"gitlab.com/CoiaPrant/lib/net/diagnostics/tcping"
	"golang.org/x/net/proxy"
)

func TestDial(t *testing.T) {
	if !tcping.Check("127.0.0.1:8080") {
		t.Skip("skipping test because the HTTP proxy is not running on 127.0.0.1:8080")
	}

	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		t.Fatalf("failed to parse url, error: %s", err)
		return
	}

	dialer, err := proxy.FromURL(u, proxy.Direct)
	if err != nil {
		t.Fatalf("failed to create dialer, error: %s", err)
		return
	}

	conn, err := dialer.Dial("tcp", "1.1.1.1:443")
	if err != nil {
		t.Fatalf("failed to dial, error: %s", err)
		return
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{ServerName: "1.1.1.1"})
	defer tlsConn.Close()

	err = tlsConn.Handshake()
	if err != nil {
		t.Fatalf("failed to handshake tls, error: %s", err)
		return
	}
}
