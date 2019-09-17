package eph

import (
	"fmt"
	"github.com/gman0/eph/pkg/layout"
)

func checkTargetAndBaseDirs(target, base string) error {
	if isNotExist, err := layout.DirectoryShouldExist(target); err != nil {
		if isNotExist {
			return fmt.Errorf("target path %s does not exist", target)
		}
		return err
	}

	if isNotExist, err := layout.DirectoryShouldExist(base); err != nil {
		if isNotExist {
			return fmt.Errorf("eph root %s does not exist", base)
		}
		return err
	}

	return nil
}
