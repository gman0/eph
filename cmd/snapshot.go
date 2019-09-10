package cmd

import (
	"fmt"
	"github.com/gman0/eph/pkg/eph"
	"github.com/spf13/cobra"
	"os"
)

var (
	Snapshot = cobra.Command{
		Use:   "snapshot",
		Short: "manage ramdisk snapshots",
		Long: `
manage ramdisk snapshots

Important: 'squashfs-tools' must be installed on the system and accessible
           from $PATH in order for snapshots to function
`,
		Example: `
# Create a new snapshot of /foo/bar
# CAUTION: make sure no writes to /foo/bar occur during snapshotting
eph snapshot new /foo/bar

# Apply the snapshot, restoring the state of the ramdisk
# to the time when the snapshot was taken
eph snapshot apply /foo/bar --id 1
`,
	}

	snapshotNew = cobra.Command{
		Use:   "new PATH",
		Short: "create a new snapshot",
		Long: `
create a new snapshot

Create a snapshot of the current state of the ramdisk.

Note that the snapshot is stored inside the ramdisk and
contributes to the overall allocated space.

Important: make sure no writes occur to the ramdisk while
           the snapshot is being taken.
           Doing so may corrupt the snapshot.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.NewSnapshot(stripTrailingSlash(args[0]), snapshotNewLabel, snapshotNewCompressionAlg, snapshotNewOnline); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	snapshotDelete = cobra.Command{
		Use:   "delete PATH -i SNAPSHOT-ID",
		Short: "delete a snapshot",
		Long: `
delete a snapshot

The snapshot being deleted must not be currently in use
and must not be a parent of any other snapshot(s).
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.DeleteSnapshot(stripTrailingSlash(args[0]), snapshotId); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	snapshotList = cobra.Command{
		Use:   "list PATH",
		Short: "list snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.PrintSnapshotsList(stripTrailingSlash(args[0])); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	snapshotShow = cobra.Command{
		Use:   "show PATH -i SNAPSHOT-ID",
		Short: "show snapshot details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.PrintSnapshotDetails(stripTrailingSlash(args[0]), snapshotId); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	snapshotApply = cobra.Command{
		Use:   "apply PATH -i SNAPSHOT-ID",
		Short: "apply a snapshot to the ramdisk",
		Long: `
apply a snapshot to the ramdisk

Ramdisk is overlayed on top of the snapshot.
Any existing data that's currently stored in the ramdisk is discarded.

This operation requires overlay remount.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPathArg(args); err != nil {
				return err
			}

			if err := eph.ApplySnapshot(stripTrailingSlash(args[0]), snapshotId); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	snapshotNewLabel          string
	snapshotNewCompressionAlg string
	snapshotNewOnline         bool

	snapshotId int
)

func init() {
	Snapshot.AddCommand(&snapshotNew)
	Snapshot.AddCommand(&snapshotDelete)
	Snapshot.AddCommand(&snapshotApply)
	Snapshot.AddCommand(&snapshotList)
	Snapshot.AddCommand(&snapshotShow)

	snapshotNew.PersistentFlags().StringVarP(&snapshotNewLabel, "label", "l", "", "snapshot label")
	snapshotNew.PersistentFlags().StringVarP(&snapshotNewCompressionAlg, "compression", "c", "xz", "compression algorithm to use for the new snapshot; available gzip, lzo, xz")

	snapshotDelete.PersistentFlags().IntVarP(&snapshotId, "id", "i", 0, "snapshot ID")
	snapshotDelete.MarkPersistentFlagRequired("id")

	snapshotApply.PersistentFlags().IntVarP(&snapshotId, "id", "i", 0, "snapshot ID")
	snapshotApply.MarkPersistentFlagRequired("id")

	snapshotShow.PersistentFlags().IntVarP(&snapshotId, "id", "i", 0, "snapshot ID")
	snapshotShow.MarkPersistentFlagRequired("id")
}
