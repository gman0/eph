package onerror

import "fmt"

func wrapErr(msg []string, err error) error {
	if err != nil {
		if len(msg) > 0 {
			return fmt.Errorf("%s: %v", msg[0], err)
		}

		return err
	}

	return nil
}
