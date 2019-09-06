package layout

import (
	"fmt"
	"os"
)

func DirectoryShouldExist(p string) error {
	if st, err := os.Lstat(p); err != nil {
		return err
	} else {
		if !st.IsDir() {
			return fmt.Errorf("%s is not a directory", p)
		}
	}

	return nil
}

func PathShouldNotExist(p string) error {
	if _, err := os.Lstat(p); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		return fmt.Errorf("%s already exists", p)
	}

	return nil
}
