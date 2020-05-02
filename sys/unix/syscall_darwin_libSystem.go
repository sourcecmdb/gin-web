package unix

import (
	"syscall"
	"unsafe"
)

func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)

func funcPC(f func()) uintptr {
	return **(**uintptr)(unsafe.Pointer(&f))
}
