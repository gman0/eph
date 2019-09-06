package device

import (
	"fmt"
	"golang.org/x/sys/unix"
	"strings"
)

func Overlay(into, upperDir, workDir string, lowerDirs ...string) error {
	lowerOpts := strings.Join(lowerDirs, ":")
	return unix.Mount("overlay", into, "overlay", 0,
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerOpts, upperDir, workDir))
}
