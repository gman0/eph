package layout

import (
	"fmt"
	"path"
)

const (
	fmtBase = ".eph.%s"
	fmtOrig = "%s/orig"

	fmtStaging        = "%s/staging"
	fmtOverlayHead    = "%s/staging/head"
	fmtOverlayDiff    = "%s/staging/diff"
	fmtOverlayWorkdir = "%s/staging/work"

	fmtSnapshots      = "%s/staging/snapshots"
	fmtSnapshotsState = "%s/staging/snapshots/state"
	fmtSnapshotMounts = "%s/staging/snapshots/mounts"
)

var (
	BaseOverride string
)

func fmtPath(format, p string) string {
	return fmt.Sprintf(format, Base(p))
}

func Base(p string) string {
	if BaseOverride != "" {
		return BaseOverride
	}
	return path.Join(path.Dir(p), fmt.Sprintf(fmtBase, path.Base(p)))
}

func Orig(p string) string { return fmtPath(fmtOrig, p) }

func Staging(p string) string { return fmtPath(fmtStaging, p) }

func Head(p string) string { return fmtPath(fmtOverlayHead, p) }

func OverlayDiff(p string) string { return fmtPath(fmtOverlayDiff, p) }

func OverlayWorkdir(p string) string { return fmtPath(fmtOverlayWorkdir, p) }

func Snapshots(p string) string { return fmtPath(fmtSnapshots, p) }

func SnapshotsState(p string) string { return fmtPath(fmtSnapshotsState, p) }

func SnapshotMounts(p string) string { return fmtPath(fmtSnapshotMounts, p) }
