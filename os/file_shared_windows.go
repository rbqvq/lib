package os

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"unsafe"

	libsyscall "gitlab.com/CoiaPrant/lib/os/internal/syscall"
	"golang.org/x/sys/windows"
)

var errSetAppendMode = errors.New("failed to set append mode on file")

func setAppendMode(f *os.File, appendMode bool) bool {
	if f == nil {
		return false
	}

	file := reflect.ValueOf(f).Elem().FieldByName("file") // *os.file
	if file.IsNil() {
		return false
	}

	field := file.Elem().FieldByName("appendMode") // (*os.file).appendMode
	if !field.IsValid() || field.Kind() != reflect.Bool {
		return false
	}

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().SetBool(appendMode)
	return true
}

func openat(dirfd windows.Handle, name string, flag uint64, perm os.FileMode) (windows.Handle, error) {
	return libsyscall.Openat(dirfd, name, flag|windows.O_CLOEXEC, syscallMode(perm))
}

// OpenFileShared is the generalized form of OpenShared.
// It opens the named file with the given flags and permissions,
// following the same semantics as OpenFile, except that the file
// remains replaceable and removable while it is open and is opened
// relative to a handle on its parent directory.
func OpenFileShared(name string, flag int, perm os.FileMode) (*os.File, error) {
	parent, err := os.OpenFile(filepath.Dir(name), os.O_RDONLY|libsyscall.O_DIRECTORY, 0)
	if err != nil {
		return nil, err
	}
	defer parent.Close()

	h, err := openat(
		windows.Handle(parent.Fd()),
		filepath.Base(name),
		uint64(flag),
		perm,
	)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}

	f := os.NewFile(uintptr(h), name)
	if flag&os.O_APPEND == 0 {
		return f, nil
	}

	if setAppendMode(f, true) {
		return f, nil
	}

	defer f.Close()
	return nil, &os.PathError{Op: "open", Path: name, Err: errSetAppendMode}
}

// OpenShared opens the named file for reading.
func OpenShared(name string) (*os.File, error) {
	return OpenFileShared(name, os.O_RDONLY, 0)
}
