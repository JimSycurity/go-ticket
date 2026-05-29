//go:build windows

package ticket

import (
	"syscall"
	"unsafe"
)

const (
	movefileReplaceExisting = 0x1
	movefileWriteThrough    = 0x8
)

func replaceFile(src string, dst string) error {
	srcp, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstp, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	moveFileEx := kernel32.NewProc("MoveFileExW")
	ret, _, err := moveFileEx.Call(
		uintptr(unsafe.Pointer(srcp)),
		uintptr(unsafe.Pointer(dstp)),
		uintptr(movefileReplaceExisting|movefileWriteThrough),
	)
	if ret == 0 {
		if err != syscall.Errno(0) {
			return err
		}
		return syscall.EINVAL
	}
	return nil
}
