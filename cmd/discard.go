package cmd

import (
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/spf13/cobra"
	"os"
)

var (
	Discard = cobra.Command{
		Use:   "discard PATH",
		Short: "discard ramdisk and unmount",
		Long: `
discard a ramdisk in specified path and restore the original data

All ramdisk data is irrevertably discarded and thrown away.
The original data restored to its initial location and the ramdisk is unmounted.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.DiscardEphemeral(args[0]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}
)
