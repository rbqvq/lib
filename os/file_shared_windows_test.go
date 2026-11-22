//go:build windows

package os

import (
	"os"
	"testing"
)

func TestSetAppendMode(t *testing.T) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if _, err := f.WriteAt([]byte("x"), 0); err != nil {
		t.Fatalf("WriteAt before toggle failed: %v", err)
	}

	if !setAppendMode(f, true) {
		t.Fatal("setAppendMode(f, true) returned false")
	}
	if _, err := f.WriteAt([]byte("y"), 0); err == nil {
		t.Fatal("WriteAt after setAppendMode(true) succeeded, want error")
	}

	if !setAppendMode(f, false) {
		t.Fatal("setAppendMode(f, false) returned false")
	}
	if _, err := f.WriteAt([]byte("z"), 0); err != nil {
		t.Fatalf("WriteAt after setAppendMode(false) failed: %v", err)
	}
}
