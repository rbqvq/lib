package quic

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gitlab.com/CoiaPrant/lib/dissectors/tls"
	"gitlab.com/go-extension/bytes"
)

func newPacketBuffer(file ...string) ([][]byte, error) {
	b := make([][]byte, 0, len(file))

	for _, f := range file {
		p, err := os.ReadFile("testdata/" + f)
		if err != nil {
			return nil, err
		}

		b = append(b, p)
	}

	return b, nil
}

func parseInitialClientHello(b [][]byte) (*tls.ClientHello, error) {
	if len(b) < 1 {
		return nil, fmt.Errorf("no packet data provided")
	}

	packet := b[0]
	b = b[1:]

	v, err := ParseVersion(packet)
	if err != nil {
		return nil, err
	}

	if !v.IsSupportedVersion() {
		return nil, ErrNotQUIC
	}

	initialPacket, err := ParseInitialPacket(packet, WithVersion(v))
	if err != nil {
		return nil, err
	}

	var retry int
	frames := initialPacket.Frames
	crypto, ok := AssembleCryptoFrames(frames)
	for !ok {
		if retry > 4 {
			return nil, fmt.Errorf("failed to assemble crypto frames after %d retries", retry)
		}
		if len(b) < 1 {
			return nil, fmt.Errorf("no packet data provided")
		}

		packet = b[0]
		b = b[1:]

		initialPacket, err = ParseInitialPacket(packet, WithInitialPacket(initialPacket))
		if err != nil {
			return nil, err
		}

		frames = append(frames, initialPacket.Frames...)
		crypto, ok = AssembleCryptoFrames(frames)
		retry++
	}

	if !v.TLS() {
		msg, err := ParseCryptoGoogle(crypto)
		if err != nil {
			return nil, err
		}

		if msg.Tag != "CHLO" {
			return nil, fmt.Errorf("invalid google quic tag: %s", msg.Tag)
		}

		var clientHello = &tls.ClientHello{}
		if value, ok := msg.TagValues["SNI\x00"]; ok {
			clientHello.ServerName = strings.ToLower(string(value))
		}
		return clientHello, nil
	}

	return tls.ParseHandshakeClientHello(bytes.NewReader(crypto))
}

func decryptInitialClientHello(t *testing.T, name string, file ...string) {
	t.Run(name, func(t *testing.T) {
		b, err := newPacketBuffer(file...)
		if err != nil {
			t.Fatalf("readfile failed: %s", err)
			return
		}

		clientHello, err := parseInitialClientHello(b)
		if err != nil {
			t.Fatalf("quic initial decrypt failed: %s", err)
			return
		}

		t.Logf("quic initial decrypted, sni: %s, alpn: %s", clientHello.ServerName, clientHello.ALPN)
	})
}

func TestDecryptInitialClientHello(t *testing.T) {
	tests := []struct {
		name string
		file []string
	}{
		{"Q043", []string{"quic-q043.bin"}},
		{"Q046", []string{"quic-q046.bin"}},
		{"Q050", []string{"quic-q050.bin"}},
		{"T050", []string{"quic-t050.bin"}},
		{"T051", []string{"quic-t051.bin"}},
		{"V1_Draft27", []string{"quic-v1-draft-27.bin"}},
		{"V1_Draft29", []string{"quic-v1-draft-29.bin"}},
		{"V1", []string{"quic-v1.bin"}},
		{"V1_Scramble", []string{"quic-v1-scramble-1.bin", "quic-v1-scramble-2.bin"}},
		{"V1_Retransmission", []string{"quic-v1-retransmission-2.bin", "quic-v1-retransmission-3.bin"}},
	}

	for _, tt := range tests {
		decryptInitialClientHello(t, tt.name, tt.file...)
	}
}
