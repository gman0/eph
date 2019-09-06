package cmd

import (
	"errors"
	"regexp"
)

var (
	quotaRegexp = regexp.MustCompile(`\d+[KMGT]`)
)

func checkPathArg(args []string) error {
	if len(args) == 0 {
		return errors.New("missing path")
	}

	if len(args) != 1 {
		return errors.New("expected exactly one path argument")
	}

	return nil
}

func checkQuotaFormat(quota string) bool {
	return quotaRegexp.MatchString(quota)
}
