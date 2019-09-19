package device

import (
	"golang.org/x/sys/unix"
	"syscall"
)

func BindRW(from, to string) error {
	return unix.Mount(from, to, "", syscall.MS_BIND, "")
}

func BindRO(from, to string) error {
	if err := BindRW(from, to); err != nil {
		return err
	}

	return unix.Mount(from, to, "", syscall.MS_REMOUNT|syscall.MS_BIND|syscall.MS_RDONLY, "")
}
