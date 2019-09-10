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

func stripTrailingSlash(p string) string {
	if len(p) > 1 {
		if p[len(p)-1] == '/' {
			return p[:len(p)-1]
		}
	}
	return p
}

func checkQuotaFormat(quota string) bool {
	return quotaRegexp.MatchString(quota)
}
