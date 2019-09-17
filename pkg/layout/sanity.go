package layout

import (
	"fmt"
	"os"
)

func DirectoryShouldExist(p string) (isNotExist bool, err error) {
	if st, err := os.Lstat(p); err != nil {
		return os.IsNotExist(err), err
	} else {
		if !st.IsDir() {
			return false, fmt.Errorf("%s is not a directory", p)
		}
	}

	return false, nil
}

func PathShouldNotExist(p string) (exists bool, err error) {
	if _, err := os.Lstat(p); err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
	} else {
		return true, fmt.Errorf("%s already exists", p)
	}

	return false, nil
}
