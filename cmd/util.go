package cmd

import (
	"errors"
	"github.com/gman0/eph/pkg/layout"
	"regexp"
)

var (
	quotaRegexp = regexp.MustCompile(`\d+[KMG]`)
)

func checkPathArg(args []string) error {
	if len(args) == 0 {
		return errors.New("missing path")
	}

	if len(args) != 1 {
		return errors.New("expected exactly one path argument")
	}

	if args[0] == layout.BaseOverride {
		return errors.New("eph root collision")
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
