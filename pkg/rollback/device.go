package rollback

import "github.com/gman0/eph/pkg/device"

func Unmount(mountPoint string, err *error) error {
	if *err != nil {
		return device.Unmount(mountPoint)
	}
	return nil
}
