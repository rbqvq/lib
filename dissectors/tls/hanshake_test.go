package tls

import (
	"fmt"
	"io"
	"os"
	"testing"

	"gitlab.com/go-extension/bytes"
)

const (
	maxPlaintext = 16384
	maxHandshake = 65536 // maximum handshake we support (protocol max is 16 MB)
)

func readClientHelloRecord(b []byte) (*ClientHello, error) {
	var handshakeData []byte
	for {
		if len(handshakeData) > maxHandshake {
			return nil, fmt.Errorf("handshake too long exceeding %d bytes", maxHandshake)
		}

		if len(b) < 5 {
			return nil, fmt.Errorf("record too short: %d", len(b))
		}

		hdr := b[:5]
		b = b[5:]

		h, err := ParseRecordHeader(hdr)
		if err != nil {
			return nil, err
		}

		if h.Type != RecordTypeHandshake {
			return nil, fmt.Errorf("bad record type: %d", h.Type)
		}

		if h.Length == 0 || h.Length > maxPlaintext {
			return nil, fmt.Errorf("bad record length: %d", h.Length)
		}

		if len(b) < int(h.Length) {
			return nil, fmt.Errorf("record too short: %d", len(b))
		}

		switch handshakeData {
		case nil:
			handshakeData = b[0:h.Length:h.Length]
		default:
			handshakeData = append(handshakeData, b[:h.Length]...)
		}
		b = b[h.Length:]

		msg, err := ParseHandshakeClientHello(bytes.NewReader(handshakeData))
		if err != io.ErrUnexpectedEOF {
			return msg, err
		}
	}
}

func newBuffer(filename string) ([]byte, error) {
	return os.ReadFile("testdata/" + filename)
}

func TestParseClientHello(t *testing.T) {
	b, err := newBuffer("tls-clienthello.bin")
	if err != nil {
		t.Fatalf("readfile failed, error: %s", err)
		return
	}

	clientHello, err := readClientHelloRecord(b)
	if err != nil {
		t.Fatalf("tls clientHello parse failed, error: %s", err)
		return
	}

	fmt.Printf("tls clientHello parsed, sni: %s, alpn: %s\n", clientHello.ServerName, clientHello.ALPN)
}
