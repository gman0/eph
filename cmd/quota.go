package cmd

import (
	"errors"
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/spf13/cobra"
	"os"
)

var (
	SetQuota = cobra.Command{
		Use:   "set-quota PATH -q SIZE",
		Short: "set ramdisk quota",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if setQuota == "" {
				return errors.New("missing quota")
			}

			if !checkQuotaFormat(setQuota) {
				return errors.New("invalid quota format")
			}

			if err := eph.SetQuota(args[0], setQuota); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	setQuota string
)

func init() {
	SetQuota.PersistentFlags().StringVarP(&setQuota, "quota", "q", "", "quota; accepts K,M,G,T units (e.g. 1G)")
	SetQuota.MarkPersistentFlagRequired("quota")
}
