// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libsyscall

import (
	"io/fs"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// Openat flags supported by windows.Open.
const (
	O_DIRECTORY = 0x04000 // target must be a directory
)

// Openat flags not supported by windows.Open.
//
// These are invented values, use values in the 33-63 bit range
// to avoid overlap with flags and attributes supported by [windows.Open].
//
// When adding a new flag here, add an unexported version to
// the set of invented O_ values in windows/types_windows.go
// to avoid overlap.
const (
	O_NOFOLLOW_ANY = 0x200000000 // disallow symlinks anywhere in the path
	O_WRITE_ATTRS  = 0x800000000 // FILE_WRITE_ATTRIBUTES, used by Chmod
)

func Openat(dirfd windows.Handle, name string, flag uint64, perm uint32) (_ windows.Handle, e1 error) {
	if len(name) == 0 {
		return windows.InvalidHandle, windows.ERROR_FILE_NOT_FOUND
	}

	var access, options uint32
	// Map Win32 file flags to NT create options.
	fileFlags := uint32(flag) & FileFlagsMask
	if fileFlags&^ValidFileFlagsMask != 0 {
		return windows.InvalidHandle, fs.ErrInvalid
	}
	if fileFlags&windows.O_FILE_FLAG_OVERLAPPED == 0 {
		options |= windows.FILE_SYNCHRONOUS_IO_NONALERT
	}
	if fileFlags&windows.O_FILE_FLAG_DELETE_ON_CLOSE != 0 {
		access |= windows.DELETE
	}
	setOptionFlag := func(ntFlag, win32Flag uint32) {
		if fileFlags&win32Flag != 0 {
			options |= ntFlag
		}
	}
	setOptionFlag(windows.FILE_NO_INTERMEDIATE_BUFFERING, windows.O_FILE_FLAG_NO_BUFFERING)
	setOptionFlag(windows.FILE_WRITE_THROUGH, windows.O_FILE_FLAG_WRITE_THROUGH)
	setOptionFlag(windows.FILE_SEQUENTIAL_ONLY, windows.O_FILE_FLAG_SEQUENTIAL_SCAN)
	setOptionFlag(windows.FILE_RANDOM_ACCESS, windows.O_FILE_FLAG_RANDOM_ACCESS)
	setOptionFlag(windows.FILE_OPEN_FOR_BACKUP_INTENT, windows.O_FILE_FLAG_BACKUP_SEMANTICS)
	setOptionFlag(FILE_SESSION_AWARE, windows.O_FILE_FLAG_SESSION_AWARE)
	setOptionFlag(windows.FILE_DELETE_ON_CLOSE, windows.O_FILE_FLAG_DELETE_ON_CLOSE)
	setOptionFlag(windows.FILE_OPEN_NO_RECALL, windows.O_FILE_FLAG_OPEN_NO_RECALL)
	setOptionFlag(windows.FILE_OPEN_REPARSE_POINT, windows.O_FILE_FLAG_OPEN_REPARSE_POINT)

	switch flag & (windows.O_RDONLY | windows.O_WRONLY | windows.O_RDWR) {
	case windows.O_RDONLY:
		// FILE_GENERIC_READ includes FILE_LIST_DIRECTORY.
		access |= windows.FILE_GENERIC_READ
	case windows.O_WRONLY:
		access |= windows.FILE_GENERIC_WRITE
		options |= windows.FILE_NON_DIRECTORY_FILE
	case windows.O_RDWR:
		access |= windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE
		options |= windows.FILE_NON_DIRECTORY_FILE
	default:
		// Stat opens files without requesting read or write permissions,
		// but we still need to request SYNCHRONIZE.
		access |= windows.SYNCHRONIZE
	}
	if flag&windows.O_CREAT != 0 {
		access |= windows.FILE_GENERIC_WRITE
	}
	if fileFlags&windows.O_FILE_FLAG_NO_BUFFERING != 0 {
		// Disable buffering implies no implicit append access.
		access &^= windows.FILE_APPEND_DATA
	}
	if flag&windows.O_APPEND != 0 {
		access |= windows.FILE_APPEND_DATA
		// Remove FILE_WRITE_DATA access unless O_TRUNC is set,
		// in which case we need it to truncate the file.
		if flag&windows.O_TRUNC == 0 {
			access &^= windows.FILE_WRITE_DATA
		}
	}
	if flag&O_DIRECTORY != 0 {
		options |= windows.FILE_DIRECTORY_FILE
		access |= windows.FILE_LIST_DIRECTORY
	}
	if flag&windows.O_SYNC != 0 {
		options |= windows.FILE_WRITE_THROUGH
	}
	if flag&O_WRITE_ATTRS != 0 {
		access |= windows.FILE_WRITE_ATTRIBUTES
	}
	// Allow File.Stat.
	access |= windows.STANDARD_RIGHTS_READ | windows.FILE_READ_ATTRIBUTES | windows.FILE_READ_EA

	objAttrs := &windows.OBJECT_ATTRIBUTES{}
	if flag&O_NOFOLLOW_ANY != 0 {
		objAttrs.Attributes |= windows.OBJ_DONT_REPARSE
	}
	if flag&windows.O_CLOEXEC == 0 {
		objAttrs.Attributes |= windows.OBJ_INHERIT
	}
	if fileFlags&windows.O_FILE_FLAG_POSIX_SEMANTICS == 0 {
		objAttrs.Attributes |= windows.OBJ_CASE_INSENSITIVE
	}
	if err := initObjectAttributes(objAttrs, dirfd, name); err != nil {
		return windows.InvalidHandle, err
	}

	// We don't use FILE_OVERWRITE/FILE_OVERWRITE_IF, because when opening
	// a file with FILE_ATTRIBUTE_READONLY these will replace an existing
	// file with a new, read-only one.
	//
	// Instead, we ftruncate the file after opening when O_TRUNC is set.
	var disposition uint32
	switch {
	case flag&(windows.O_CREAT|windows.O_EXCL) == (windows.O_CREAT | windows.O_EXCL):
		disposition = windows.FILE_CREATE
		options |= windows.FILE_OPEN_REPARSE_POINT // don't follow symlinks
	case flag&windows.O_CREAT == windows.O_CREAT:
		disposition = windows.FILE_OPEN_IF
	default:
		disposition = windows.FILE_OPEN
	}

	fileAttrs := uint32(windows.FILE_ATTRIBUTE_NORMAL)
	if perm&windows.S_IWRITE == 0 {
		fileAttrs = windows.FILE_ATTRIBUTE_READONLY
	}

	var h windows.Handle
	var err error
	if TestOpenatFallback && flag&O_NOFOLLOW_ANY != 0 {
		err = windows.STATUS_INVALID_PARAMETER
	} else {
		err = windows.NtCreateFile(
			&h,
			windows.SYNCHRONIZE|access,
			objAttrs,
			&windows.IO_STATUS_BLOCK{},
			nil,
			fileAttrs,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
			disposition,
			windows.FILE_OPEN_FOR_BACKUP_INTENT|options,
			0,
			0,
		)
	}
	if err == windows.STATUS_INVALID_PARAMETER && flag&O_NOFOLLOW_ANY != 0 {
		h, err = openatFallback(name, windows.SYNCHRONIZE|access, *objAttrs, fileAttrs, disposition, windows.FILE_OPEN_FOR_BACKUP_INTENT|options)
	}
	if err != nil {
		return h, ntCreateFileError(err, flag)
	}

	if flag&windows.O_TRUNC != 0 {
		err = windows.Ftruncate(h, 0)
		if err == windows.ERROR_INVALID_PARAMETER {
			// ERROR_INVALID_PARAMETER means truncation is not supported on this file handle.
			// Unix's O_TRUNC specification says to ignore O_TRUNC on named pipes and terminal devices.
			// We do the same here.
			if t, err1 := windows.GetFileType(h); err1 == nil && (t == windows.FILE_TYPE_PIPE || t == windows.FILE_TYPE_CHAR) {
				err = nil
			}
		}
		if err != nil {
			windows.CloseHandle(h)
			return windows.InvalidHandle, err
		}
	}

	return h, nil
}

// TestOpenatFallback should only be used for testing purposes.
// When set, Openat simulates a system that does not support OBJ_DONT_REPARSE.
var TestOpenatFallback bool

// openatFallback implements O_NOFOLLOW_ANY for a single path component on
// Windows versions that do not support OBJ_DONT_REPARSE (including Windows 10
// build 10240). See go.dev/issue/78131.
func openatFallback(name string, access uint32, attrs windows.OBJECT_ATTRIBUTES, fileAttrs, disposition, options uint32) (windows.Handle, error) {
	// FILE_OPEN_REPARSE_POINT only prevents following the final component.
	// All current production callers that request O_NOFOLLOW_ANY pass a
	// single component: os.Root splits paths (including symlink targets)
	// before opening each component, and RemoveAll passes a basename or a
	// directory entry name relative to an open parent directory. Thus this
	// fallback also supports multi-component paths passed to those APIs.
	// Reject other paths here rather than weaken O_NOFOLLOW_ANY: this is
	// not a general replacement for OBJ_DONT_REPARSE.
	if attrs.RootDirectory == 0 || name == ".." ||
		strings.IndexByte(name, '\\') >= 0 || strings.IndexByte(name, '/') >= 0 || strings.IndexByte(name, ':') >= 0 {
		return windows.InvalidHandle, windows.STATUS_INVALID_PARAMETER
	}
	attrs.Attributes &^= windows.OBJ_DONT_REPARSE
	var h windows.Handle
	err := windows.NtCreateFile(
		&h, access, &attrs, &windows.IO_STATUS_BLOCK{}, nil, fileAttrs,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, disposition,
		(options|windows.FILE_OPEN_REPARSE_POINT)&^windows.FILE_DELETE_ON_CLOSE, 0, 0,
	)
	if err != nil {
		return windows.InvalidHandle, err
	}
	// Inspect the handle, not the path, before truncation or delete-on-close.
	// Skip this check if opening the reparse point itself was requested,
	// or if O_CREAT|O_EXCL requires creating a new file without following links.
	if options&windows.FILE_OPEN_REPARSE_POINT == 0 {
		var info windows.ByHandleFileInformation
		err = windows.GetFileInformationByHandle(h, &info)
		if err == nil && info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			err = windows.STATUS_REPARSE_POINT_ENCOUNTERED
		}
		if err != nil {
			windows.CloseHandle(h)
			return windows.InvalidHandle, err
		}
	}
	if options&windows.FILE_DELETE_ON_CLOSE == 0 {
		return h, nil
	}
	defer windows.CloseHandle(h)

	// Only enable deletion after checking for a reparse point. An empty
	// name relative to h reopens the same file, even if it has been renamed
	// or its directory entry has been replaced since the first open.
	// The file already exists, including when created with O_EXCL.
	attrs.RootDirectory = h
	attrs.ObjectName = &windows.NTUnicodeString{}
	var dh windows.Handle
	err = NtOpenFile(
		&dh, access, &attrs, &windows.IO_STATUS_BLOCK{},
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		options|windows.FILE_OPEN_REPARSE_POINT,
	)
	if err != nil {
		return windows.InvalidHandle, err
	}
	return dh, nil
}

// ntCreateFileError maps error returns from NTCreateFile to user-visible errors.
func ntCreateFileError(err error, flag uint64) error {
	s, ok := err.(windows.NTStatus)
	if !ok {
		// The Openat fallback can also return Win32 errors.
		return err
	}
	switch s {
	case windows.STATUS_REPARSE_POINT_ENCOUNTERED:
		return syscall.ELOOP
	case windows.STATUS_NOT_A_DIRECTORY:
		// ENOTDIR is the errno returned by open when O_DIRECTORY is specified
		// and the target is not a directory.
		//
		// NtCreateFile can return STATUS_NOT_A_DIRECTORY under other circumstances,
		// such as when opening "file/" where "file" is not a directory.
		// (This might be Windows version dependent.)
		//
		// Only map STATUS_NOT_A_DIRECTORY to ENOTDIR when O_DIRECTORY is specified.
		if flag&O_DIRECTORY != 0 {
			return syscall.ENOTDIR
		}
	case windows.STATUS_FILE_IS_A_DIRECTORY:
		return syscall.EISDIR
	case windows.STATUS_OBJECT_NAME_COLLISION:
		return syscall.EEXIST
	}
	return s.Errno()
}
