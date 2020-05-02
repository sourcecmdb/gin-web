package go_isatty

import "github.com/sourcecmdb/gin-web/sys/unix"

// 如果文件描述符为终端，则IsTerminal返回true IsTerminal return true if the file descriptor is terminal
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd),unix.TIOCGETA)
	if err == nil
}

func IsCygwinTerminal(fd uintptr)bool{
	return false
}
