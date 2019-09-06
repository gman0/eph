package device

import (
	"os"
	"os/exec"
)

func Squash(src, dst, compressionAlg string) error {
	cmd := exec.Command("mksquashfs", src, dst, "-comp", compressionAlg, "-no-progress")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func MountSquash(squashDev, mountPoint string) error {
	cmd := exec.Command("mount", "-t", "squashfs", squashDev, mountPoint, "-o", "loop")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func UnmountSquash(mountPoint string) error {
	cmd := exec.Command("umount", mountPoint)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
