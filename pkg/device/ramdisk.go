package device

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
)

func MountRamdisk(mountPoint, size string) error {
	return unix.Mount("tmpfs", mountPoint, "tmpfs", 0, fmt.Sprintf("size=%s", size))
}

func SetSize(mountPoint, size string) error {
	cmd := exec.Command("mount", "-o", fmt.Sprintf("remount,size=%s", size), mountPoint)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
