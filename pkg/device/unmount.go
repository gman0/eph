package device

import "golang.org/x/sys/unix"

func Unmount(mountPoint string) error {
	return unix.Unmount(mountPoint, 0)
}
