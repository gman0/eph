package rollback

import "os"

func Remove(p string, err *error) error {
	if *err != nil {
		return os.Remove(p)
	}
	return nil
}

func Rename(oldPath, newPath string, err *error) error {
	if *err != nil {
		return os.Rename(oldPath, newPath)
	}
	return nil
}
