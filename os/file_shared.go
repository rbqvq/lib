//go:build !windows

package os

import "os"

// OpenShared opens the named file for reading.
//
// On Unix, this is equivalent to Open: POSIX unlink semantics already allow
// the file to be replaced or removed while open. Unlike the Windows
// implementation, the path is resolved in full and is not anchored to a
// parent directory handle, so a concurrent rename of a parent directory may
// cause a different file to be opened.
func OpenShared(name string) (*os.File, error) {
	return os.Open(name)
}

// OpenShared opens the named file for reading.
func OpenFileShared(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}
