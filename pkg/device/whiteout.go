package device

import (
	"os"
	"syscall"
)

func IsWhiteout(info os.FileInfo) bool {
	st := info.Sys().(*syscall.Stat_t)
	return st.Rdev == 0 && (st.Mode&syscall.S_IFMT) == syscall.S_IFCHR
}
