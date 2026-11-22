package libsyscall

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go mksyscall_windows.go

//sys   NtOpenFile(handle *windows.Handle, access uint32, oa *windows.OBJECT_ATTRIBUTES, iosb *windows.IO_STATUS_BLOCK, share uint32, options uint32) (ntstatus error) = ntdll.NtOpenFile
