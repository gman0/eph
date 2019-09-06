package layout

import "fmt"

func SnapshotFilename(snapId int) string {
	return fmt.Sprintf("snap-%d.squash", snapId)
}

func SnapshotMountpointTarget(snapId int) string {
	return fmt.Sprintf("snap-%d.mount", snapId)
}
