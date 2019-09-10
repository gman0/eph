package cmd

import (
	"errors"
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/gman0/eph/pkg/layout"
	"github.com/spf13/cobra"
	"os"
)

var (
	Create = cobra.Command{
		Use:   "create PATH",
		Short: "create a new ramdisk",
		Long: `
create a new ramdisk in specified path

A ramdisk is a tmpfs mount.

A ramdisk may be either blank or overlay'd over an existing directory.
When creating a blank ramdisk, the specified path is expected to not exist.
You may also overlay over an existing directory with the --overlay flag.
The original file-system subtree is accessed in read-only mode, and all
modifications to the subtree are overlay'd over the subtree and stored in
the ramdisk. See OverlayFS docs for more info:
  https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt

Let /foo/BAR be the directory you want to overlay over.
eph's staging directory is created in /foo/.eph.BAR
The original /foo/BAR is moved to /foo/.eph.BAR/orig
In case of any unexpected failures, you can always retrieve the original
data from there. /foo/BAR is now bound to an OverlayFS mount point in
eph's staging directory.
`,
		Example: `
# Create a blank ramdisk in /foo. /foo must not exist
eph create /foo

# Create an overlay'd ramdisk over an existing directory 
eph create -o /bar
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if createQuota == "" {
				return errors.New("missing quota")
			}

			if !checkQuotaFormat(createQuota) {
				return errors.New("invalid quota format")
			}

			p := stripTrailingSlash(args[0])

			if !createOverlay {
				if err := layout.PathShouldNotExist(p); err != nil {
					return err
				}

				if err := os.Mkdir(p, 0755); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}

			if err := eph.Create(p, createQuota); err != nil {
				fmt.Fprintln(os.Stderr, err)

				if !createOverlay {
					os.Remove(p)
				}

				os.Exit(1)
			}

			return nil
		},
	}

	createQuota   string
	createOverlay bool
)

func init() {
	Create.PersistentFlags().StringVarP(&createQuota, "quota", "q", "100M", "ramdisk capacity quota; accepts K,M,G,T units")
	Create.PersistentFlags().BoolVarP(&createOverlay, "overlay", "o", false, "overlay over an existing directory")
}
