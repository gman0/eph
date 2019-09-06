package layout

import (
	"fmt"
	"path"
)

const (
	fmtBase = ".eph.%[1]s"
	fmtOrig = ".eph.%[1]s/orig"
	fmtHead = ".eph.%[1]s/head"

	fmtStaging        = ".eph.%[1]s/staging"
	fmtOverlayDiff    = ".eph.%[1]s/staging/diff"
	fmtOverlayWorkdir = ".eph.%[1]s/staging/work"

	fmtSnapshots      = ".eph.%[1]s/staging/snapshots"
	fmtSnapshotsState = ".eph.%[1]s/staging/snapshots/state"
	fmtSnapshotMounts = ".eph.%[1]s/staging/snapshots/mounts"
)

func fmtPath(format, p string) string {
	return path.Join(path.Dir(p), fmt.Sprintf(format, path.Base(p)))
}

func Base(p string) string { return fmtPath(fmtBase, p) }

func Orig(p string) string { return fmtPath(fmtOrig, p) }

func Head(p string) string { return fmtPath(fmtHead, p) }

func Staging(p string) string { return fmtPath(fmtStaging, p) }

func OverlayDiff(p string) string { return fmtPath(fmtOverlayDiff, p) }

func OverlayWorkdir(p string) string { return fmtPath(fmtOverlayWorkdir, p) }

func Snapshots(p string) string { return fmtPath(fmtSnapshots, p) }

func SnapshotsState(p string) string { return fmtPath(fmtSnapshotsState, p) }

func SnapshotMounts(p string) string { return fmtPath(fmtSnapshotMounts, p) }
