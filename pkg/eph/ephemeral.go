package eph

import (
	"fmt"
	"github.com/gman0/eph/pkg/device"
	"github.com/gman0/eph/pkg/diriter"
	"github.com/gman0/eph/pkg/layout"
	"github.com/gman0/eph/pkg/onerror"
	"os"
	"os/exec"
	"path"
	"syscall"
)

type changeStatusCode byte

const (
	statusSkip     changeStatusCode = 0
	statusModified                  = 'M'
	statusAdded                     = 'A'
	statusDeleted                   = 'D'
)

func Create(source, targetOverride, size string) error {
	if isNotExist, err := layout.DirectoryShouldExist(source); err != nil {
		if isNotExist {
			return fmt.Errorf("target path %s does not exist", source)
		}
		return err
	}

	var p string
	if targetOverride != "" {
		if exists, err := layout.PathShouldNotExist(targetOverride); err != nil {
			if exists {
				return fmt.Errorf("target override %s already exists", targetOverride)
			}
			return err
		}

		p = targetOverride
	} else {
		p = source
	}

	info, err := os.Lstat(source)
	if err != nil {
		return nil
	}

	var (
		orig           = layout.Orig(p)
		base           = layout.Base(p)
		staging        = layout.Staging(p)
		head           = layout.Head(p)
		overlayDiff    = layout.OverlayDiff(p)
		overlayWorkdir = layout.OverlayWorkdir(p)
		snapshotsBase  = layout.Snapshots(p)
		snapshotMounts = layout.SnapshotMounts(p)
		snapshotsState = layout.SnapshotsState(p)
	)

	if baseExists, err := layout.PathShouldNotExist(base); err != nil {
		if baseExists {
			return fmt.Errorf("eph root %s already exists", base)
		} else {
			return err
		}
	}

	wrapE := func(msg string, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %v", msg, err)
		}
		return nil
	}

	var (
		ss = SnapshotsState{}
		st = info.Sys().(*syscall.Stat_t)
	)

	do := onerror.Rollback{}

	// Prepare ramdisk
	do.
		TryMkDir(base, 0755, "failed to create eph root").
		TryMkDir(staging, 0700).
		TryMountRamdisk(staging, size, "failed to mount ramdisk").
		Try(func() error { return mkDirs(0700, head, overlayDiff, overlayWorkdir, snapshotsBase, snapshotMounts) }, func() {}).
		Try(func() error { return wrapE("failed to write snapshots state", ss.write(snapshotsState)) }, func() { os.Remove(snapshotsState) })

	if targetOverride != "" {
		do.TrySymlink(source, orig, "failed to symlink orig")
	} else {
		do.TryRename(p, orig, "failed to move orig")
	}

	// Move the original data and overlay it with the ramdisk
	do.
		TryMkDir(p, info.Mode().Perm()).
		TryBindRO(orig, head, "failed to mount HEAD").
		TryOverlayRW(p, overlayDiff, overlayWorkdir, head, "overlay mount failed").
		TryChmod(p, info.Mode()).
		TryChown(p, int(st.Uid), int(st.Gid))

	return do.Err()
}

func DiscardEphemeral(p string, noUnmount bool) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	var err error

	defer func() {
		if err != nil {
			fmt.Fprintf(os.Stderr, "discard failed:\n  recovery:\n    original data: %s\n", layout.Orig(p))
		}
	}()

	if !noUnmount {
		if err = device.Unmount(p); err != nil {
			return fmt.Errorf("failed to unmount overlay %s: %v", p, err)
		}
	}

	if err = os.Remove(p); err != nil {
		return fmt.Errorf("failed to remove overlay mount point %s: %v", p, err)
	}

	if err = destroyEph(p, noUnmount); err != nil {
		return err
	}

	return nil
}

func Merge(p string) error {
	var (
		orig = layout.Orig(p)
		base = layout.Base(p)
		diff = layout.OverlayDiff(p)
	)

	if err := checkTargetAndBaseDirs(p, base); err != nil {
		return err
	}

	if err := device.Unmount(p); err != nil {
		return fmt.Errorf("failed to unmount overlay %s: %v", p, err)
	}

	if err := os.Remove(p); err != nil {
		return fmt.Errorf("failed to remove overlay mount point %s: %v", p, err)
	}

	ss, err := readSnapshotsState(layout.SnapshotsState(p))
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	layers, err := snapshotLayers(ss, p)
	if err != nil {
		return err
	}

	cp := func(from, to string) error {
		cmd := exec.Command("cp", "--no-target-directory", "--recursive", "--no-dereference", "--preserve=all", from, to)
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	doMerge := func() error {
		for i := len(layers) - 1; i >= 0; i-- {
			iter, err := diriter.NewRecursiveIter(layers[i])
			if err != nil {
				return err
			}
			defer iter.Close()

			for !iter.AtEnd() {
				stagingPath := path.Join(iter.Base(), iter.FileInfo().Name())

				isOlderVersion, err := compareLayerVersion(stagingPath, layers, i, iter.FileInfo())
				if err != nil {
					return err
				}

				if isOlderVersion {
					iter.OrthogonalIncrement()
					continue
				}

				status, err := resolveChangeStatusCode(layers[i], orig, stagingPath, iter.FileInfo())
				if err != nil {
					return err
				}

				printStatus(iter.FileInfo(), stagingPath, layers[i], status)

				if status != statusSkip {
					origPath := orig + stagingPath[len(layers[i]):]
					switch status {
					case statusAdded:
						if iter.FileInfo().IsDir() {
							if err := device.RemoveOpaqueAttr(stagingPath); err != nil {
								return fmt.Errorf("failed to remove trusted.overlay.opaque xattr for %s: %v", stagingPath, err)
							}
						}
						fallthrough
					case statusModified:
						if err := cp(stagingPath, origPath); err != nil {
							return err
						}
					case statusDeleted:
						if err := os.RemoveAll(origPath); err != nil {
							return err
						}
					}
				}

				iter.Increment()
			}
		}

		return nil
	}

	if err := doMerge(); err != nil {
		return fmt.Errorf("merge failed, the original data may have been modified: %v\n  recovery:\n    original data: %s\n    ramdisk diff:  %s", err, orig, diff)
	}

	return destroyEph(p, false)
}

func compareLayerVersion(stagingPath string, layers []string, lowerLayerIdx int, stagingInfo os.FileInfo) (skip bool, err error) {
	relPath := stagingPath[len(layers[lowerLayerIdx]):]

	// Look for an existing dirent in higher layers
	// This is to make sure the i-th layer has the most
	// recent version of the dirent at stagingPath.
	for j := lowerLayerIdx + 1; j < len(layers); j++ {
		if _, err := os.Lstat(layers[j] + relPath); err != nil {
			if !os.IsNotExist(err) {
				return false, err
			}
		} else {
			// relPath exists in j-th layer, which means
			// the i-th layer's relPath must be older than j-th layer's.
			return true, nil
		}
	}

	return false, nil
}

func snapshotLayers(ss *SnapshotsState, p string) ([]string, error) {
	snapshotMounts := layout.SnapshotMounts(p)
	diff := layout.OverlayDiff(p)

	var layers []string
	if ss.AppliedSnapshot > 0 {
		deps, err := snapshotDependencies(ss.AppliedSnapshot, ss)
		if err != nil {
			return nil, fmt.Errorf("failed to list snapshot dependencies: %v", err)
		}

		n := len(deps) + 2

		layers = make([]string, n)

		for i := range deps {
			layers[n-i-3] = path.Join(snapshotMounts, layout.SnapshotMountpointTarget(deps[i]))
		}

		layers[n-2] = path.Join(snapshotMounts, layout.SnapshotMountpointTarget(ss.AppliedSnapshot))
		layers[n-1] = diff
	} else {
		layers = []string{diff}
	}

	return layers, nil
}

func PrintStatus(p string) error {
	var (
		orig = layout.Orig(p)
		base = layout.Base(p)
	)

	if err := checkTargetAndBaseDirs(p, base); err != nil {
		return err
	}

	ss, err := readSnapshotsState(layout.SnapshotsState(p))
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	layers, err := snapshotLayers(ss, p)
	if err != nil {
		return err
	}

	for i := len(layers) - 1; i >= 0; i-- {
		iter, err := diriter.NewRecursiveIter(layers[i])
		if err != nil {
			return err
		}
		defer iter.Close()

		for !iter.AtEnd() {
			stagingPath := path.Join(iter.Base(), iter.FileInfo().Name())

			isOlderVersion, err := compareLayerVersion(stagingPath, layers, i, iter.FileInfo())
			if err != nil {
				return err
			}

			if isOlderVersion {
				iter.OrthogonalIncrement()
				continue
			}

			status, err := resolveChangeStatusCode(layers[i], orig, stagingPath, iter.FileInfo())
			if err != nil {
				return err
			}

			printStatus(iter.FileInfo(), stagingPath, layers[i], status)

			if iter.FileInfo().IsDir() && status == statusAdded {
				iter.OrthogonalIncrement()
			} else {
				iter.Increment()
			}
		}
	}

	return nil
}

func SetQuota(p, quota string) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	return device.SetSize(layout.Staging(p), quota)
}

func printStatus(info os.FileInfo, stagingPath, basePath string, status changeStatusCode) {
	if status != statusSkip {
		if info.IsDir() {
			status ^= 0x20
		}

		fmt.Printf("%c %s\n", status, stagingPath[len(basePath)+1:])
	}
}

func destroyEph(p string, noUnmount bool) error {
	var (
		head    = layout.Head(p)
		orig    = layout.Orig(p)
		base    = layout.Base(p)
		staging = layout.Staging(p)
	)

	if !noUnmount {
		if err := device.Unmount(head); err != nil {
			return fmt.Errorf("failed to unmount HEAD %s: %v", head, err)
		}

		if err := unmountAllSnapshots(layout.SnapshotMounts(p)); err != nil {
			return err
		}

		if err := device.Unmount(staging); err != nil {
			return fmt.Errorf("failed to unmount ramdisk %s: %v", staging, err)
		}
	}

	if err := os.Remove(staging); err != nil {
		return fmt.Errorf("failed to remove ramdisk mount point %s: %v", staging, err)
	}

	origInfo, err := os.Lstat(orig)
	if err != nil {
		return err
	}

	if origInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		if err := os.Remove(orig); err != nil {
			return fmt.Errorf("failed to remove orig symlnk %s: %v", orig, err)
		}
	} else {
		if err := os.Rename(orig, p); err != nil {
			return fmt.Errorf("failed to restore orig %s: %v", orig, err)
		}
	}

	if err := os.Remove(base); err != nil {
		return fmt.Errorf("failed to remove eph root %s: %v", base, err)
	}

	return nil
}

func resolveChangeStatusCode(overlayDiff, orig, stagingPath string, stagingInfo os.FileInfo) (changeStatusCode, error) {
	if len(stagingPath) == len(overlayDiff) {
		// We want to skip the first entry which is `diff`
		// In this case, comparing the string lengths is sufficient to make sure the paths are equal
		return statusSkip, nil
	}

	var (
		relPath    = stagingPath[len(overlayDiff):]
		statusCode changeStatusCode
	)

	origInfo, statErr := os.Lstat(orig + relPath)
	if statErr != nil && !os.IsNotExist(statErr) {
		return statusSkip, statErr
	}

	if device.IsWhiteout(stagingInfo) {
		// Upper layer contains a whiteout file.
		// Two things may have caused this:
		// (a) it exists in orig, which means the file has been removed
		// (b) it doesn't exist in orig, the file existed only in ramdisk => ignore
		if os.IsNotExist(statErr) {
			statusCode = statusSkip
		} else {
			statusCode = statusDeleted
		}
	} else {
		if os.IsNotExist(statErr) {
			statusCode = statusAdded
		} else {
			if stagingInfo.IsDir() && origInfo.IsDir() {
				if stagingInfo.Mode().Perm() == origInfo.Mode().Perm() {
					statusCode = statusSkip
				}
			} else {
				statusCode = statusModified
			}
		}
	}

	return statusCode, nil
}
