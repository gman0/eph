package device

import (
	"golang.org/x/sys/unix"
)

const opaqueXAttr = "trusted.overlay.opaque"

func IsOpaque(p string) (bool, error) {
	var val [1]byte
	_, err := unix.Getxattr(p, opaqueXAttr, val[:])
	return val[0] == 'y', err
}

func RemoveOpaqueAttr(p string) error {
	return unix.Removexattr(p, opaqueXAttr)
}
