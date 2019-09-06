package eph

import (
	"fmt"
	"github.com/gman0/eph/pkg/layout"
	"os"
)

func checkTargetAndBaseDirs(target, base string) error {
	if err := layout.DirectoryShouldExist(target); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target path %s does not exist", target)
		}
		return err
	}

	if err := layout.DirectoryShouldExist(base); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("staging directory %s does not exist", base)
		}
		return err
	}

	return nil
}
