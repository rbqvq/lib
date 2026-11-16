package quic

import (
	"testing"
)

func TestAssembleCryptoFrames(t *testing.T) {
	tests := []struct {
		name     string
		frames   []FrameCrypto
		expected []byte
		ok       bool
	}{
		{
			name: "Contiguous",
			frames: []FrameCrypto{
				{Offset: 6, Data: []byte("World")},
				{Offset: 0, Data: []byte("Hello ")},
			},
			expected: []byte("Hello World"),
			ok:       true,
		},
		{
			name: "NonContiguous",
			frames: []FrameCrypto{
				{Offset: 6, Data: []byte("World")},
				{Offset: 0, Data: []byte("Hello")},
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "Redundant",
			frames: []FrameCrypto{
				{Offset: 6, Data: []byte("World")},
				{Offset: 1, Data: []byte("ello World")},
				{Offset: 0, Data: []byte("Hello ")},
			},
			expected: []byte("Hello World"),
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, ok := AssembleCryptoFrames(tt.frames)
			if ok != tt.ok {
				t.Fatalf("expected ok=%t, got %t", tt.ok, ok)
				return
			}

			if !ok {
				return
			}

			if string(data) != string(tt.expected) {
				t.Fatalf("expected data='%s', got '%s'", tt.expected, data)
				return
			}
		})
	}
}
