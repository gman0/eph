package device

import (
	"fmt"
	"golang.org/x/sys/unix"
	"strings"
)

func OverlayRW(into, upperDir, workDir string, lowerDir string) error {
	return unix.Mount("overlay", into, "overlay", 0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir))
}

func OverlayRO(into string, lowerDirs ...string) error {
	lowerOpts := strings.Join(lowerDirs, ":")
	return unix.Mount("overlay", into, "overlay", 0,
		fmt.Sprintf("lowerdir=%s", lowerOpts))
}
