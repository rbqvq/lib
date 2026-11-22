// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libsyscall

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const FileFlagsMask = 0xFFF00000

const ValidFileFlagsMask = windows.O_FILE_FLAG_OPEN_REPARSE_POINT |
	windows.O_FILE_FLAG_BACKUP_SEMANTICS |
	windows.O_FILE_FLAG_OVERLAPPED |
	windows.O_FILE_FLAG_OPEN_NO_RECALL |
	windows.O_FILE_FLAG_SESSION_AWARE |
	windows.O_FILE_FLAG_POSIX_SEMANTICS |
	windows.O_FILE_FLAG_DELETE_ON_CLOSE |
	windows.O_FILE_FLAG_SEQUENTIAL_SCAN |
	windows.O_FILE_FLAG_NO_BUFFERING |
	windows.O_FILE_FLAG_RANDOM_ACCESS |
	windows.O_FILE_FLAG_WRITE_THROUGH

const FILE_SESSION_AWARE = 0x00040000

// initObjectAttributes sets o's RootDirectory, ObjectName, and Length.
func initObjectAttributes(o *windows.OBJECT_ATTRIBUTES, root windows.Handle, name string) error {
	if name == "." {
		name = ""
	}
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return err
	}
	o.ObjectName = objectName
	if root != windows.InvalidHandle {
		o.RootDirectory = root
	}
	o.Length = uint32(unsafe.Sizeof(*o))
	return nil
}
