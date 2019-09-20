package onerror

import (
	"fmt"
	"github.com/gman0/eph/pkg/device"
	"os"
)

type Rollback struct {
	err        error
	rollbackFs []func()
}

func wrapErr(msg []string, err error) error {
	if err != nil {
		if len(msg) > 0 {
			return fmt.Errorf("%s: %v", msg[0], err)
		}

		return err
	}

	return nil
}

func (o *Rollback) Try(f func() error, rollbackF func()) *Rollback {
	if o.err == nil {
		if o.err = f(); o.err != nil {
			for i := len(o.rollbackFs) - 1; i >= 0; i-- {
				o.rollbackFs[i]()
			}
		} else {
			o.rollbackFs = append(o.rollbackFs, rollbackF)
		}
	}

	return o
}

func (o *Rollback) TryMkDir(name string, perm os.FileMode, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, os.Mkdir(name, perm)) }, func() { os.Remove(name) })
}

func (o *Rollback) TryRename(oldName, newName string, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, os.Rename(oldName, newName)) }, func() { os.Rename(newName, oldName) })
}

func (o *Rollback) TryMountRamdisk(mountPoint, size string, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, device.MountRamdisk(mountPoint, size)) }, func() { device.Unmount(mountPoint) })
}

func (o *Rollback) TryBindRO(from, to string, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, device.BindRO(from, to)) }, func() { device.Unmount(to) })
}

func (o *Rollback) TryOverlayRW(into, upperDir, workDir string, lowerDir string, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, device.OverlayRW(into, upperDir, workDir, lowerDir)) }, func() { device.Unmount(into) })
}

func (o *Rollback) TryChmod(name string, mode os.FileMode, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, os.Chmod(name, mode)) }, func() {})
}

func (o *Rollback) TryChown(name string, uid, gid int, errMsg ...string) *Rollback {
	return o.Try(func() error { return wrapErr(errMsg, os.Chown(name, uid, gid)) }, func() {})
}

func (o *Rollback) Err() error {
	return o.err
}