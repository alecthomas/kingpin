// +build linux freebsd darwin dragonfly netbsd openbsd

package kingpin

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

func guessWidth(w io.Writer) int {
	if t, ok := w.(*os.File); ok {
		fd := t.Fd()
		var dimensions [4]uint16

		if _, _, err := syscall.Syscall6(
			syscall.SYS_IOCTL,
			uintptr(fd),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(&dimensions)),
			0, 0, 0,
		); err == 0 {
			return int(dimensions[1])
		}
	}
	return 80
}
