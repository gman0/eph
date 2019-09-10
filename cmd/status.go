package cmd

import (
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/spf13/cobra"
	"os"
)

var (
	Status = cobra.Command{
		Use:   "status PATH",
		Short: "display differences between staging and original directories",
		Long: `
display differences between staging and original directories

Each line consists of status code and object path relative to the specified path.

Status codes:
* M modified
* A added
* D deleted

Lowercase M,A,D status codes are used for directories, uppercase for all non-directories.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.PrintStatus(stripTrailingSlash(args[0])); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}
)
