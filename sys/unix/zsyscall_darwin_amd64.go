package unix

func ioctl(fd int, req uint, arg uintptr) (err error) {
	_, _, e1 := syscall_syscall(funcPC(libc_iocal_trampoline), uintptr(fd), uintptr(req), uintptr(arg))
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}
func libc_iocal_trampoline()
