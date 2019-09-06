package main

import (
	"github.com/gman0/eph/cmd"
	"github.com/spf13/cobra"
	"os"
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

	completionBash := cobra.Command{
		Use:   "completion-bash",
		Short: "generate Bash completion script",
		Long: `
generate Bash completion script

To load Bash completion, run:
. <(eph completion-bash)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	completionZSH := cobra.Command{
		Use:   "completion-zsh",
		Short: "generate ZSH completion script",
		Long: `
generate ZSH completion script

To load ZSH completion, run:
. <(eph completion-zsh)
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootCmd.GenZshCompletion(os.Stdout)
		},
	}

	rootCmd.AddCommand(&cmd.Create)
	rootCmd.AddCommand(&cmd.Status)
	rootCmd.AddCommand(&cmd.Discard)
	rootCmd.AddCommand(&cmd.Merge)
	rootCmd.AddCommand(&cmd.Snapshot)
	rootCmd.AddCommand(&cmd.SetQuota)
	rootCmd.AddCommand(&completionBash)
	rootCmd.AddCommand(&completionZSH)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
