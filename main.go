package main

import (
	"errors"
	"fmt"
	"github.com/gman0/eph/cmd"
	"github.com/gman0/eph/pkg/layout"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func main() {
	rootCmd := cobra.Command{
		Use:   "eph",
		Short: "ramdisk management tool for Linux",
		Long: `
eph (ephemeral) is a ramdisk management tool for Linux
	`,
		Version: "WIP",
	}

	completion := cobra.Command{
		Use:   "completion SHELL",
		Short: "output shell completion code for the specified shell (bash or zsh)",
		Example: `
# Completion for ZSH shell
. <(eph completion zsh)

# Completion for Bash shell
. <(eph completion bash)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("no shell specified")
			}

			switch strings.ToLower(args[0]) {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			default:
				return fmt.Errorf("unknown shell %s", args[0])
			}
		},
	}

	rootCmd.AddCommand(&cmd.Create)
	rootCmd.AddCommand(&cmd.Status)
	rootCmd.AddCommand(&cmd.Discard)
	rootCmd.AddCommand(&cmd.Merge)
	rootCmd.AddCommand(&cmd.Snapshot)
	rootCmd.AddCommand(&cmd.SetQuota)
	rootCmd.AddCommand(&completion)

	rootCmd.PersistentFlags().StringVarP(&layout.BaseOverride, "eph-root", "r", "", "override default eph root location")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
