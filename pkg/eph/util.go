package eph

import (
	"fmt"
	"github.com/gman0/eph/pkg/diriter"
	"math/bits"
	"os"
	"path"
)

func removeAllIn(dir string) error {
	iter, err := diriter.NewIter(dir)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; !iter.AtEnd(); iter.Increment() {
		if err = os.RemoveAll(path.Join(dir, iter.FileInfo().Name())); err != nil {
			return err
		}
	}

	return iter.Err()
}

func coalesceStr(s string) string {
	if s == "" {
		return "<none>"
	}
	return s
}

func humanBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	}

	base := uint(bits.Len64(bytes) / 10)
	val := float64(bytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("%.1f %ciB", val, " KMGTPE"[base])
}
