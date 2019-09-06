package cmd

import (
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/spf13/cobra"
	"os"
)

var (
	Merge = cobra.Command{
		Use:   "merge PATH",
		Short: "merge ramdisk and unmount",
		Long: `
merge ramdisk and close ramdisk

Ramdisk is merged into the original data and then it's unmounted'.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.Merge(args[0]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}
)
