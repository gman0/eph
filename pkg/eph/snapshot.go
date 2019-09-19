package eph

import (
	"encoding/json"
	"fmt"
	"github.com/gman0/eph/pkg/device"
	"github.com/gman0/eph/pkg/diriter"
	"github.com/gman0/eph/pkg/layout"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type Snapshot struct {
	Id      int       `json:"id"`
	Parent  int       `json:"parent,omitempty"`
	Label   string    `json:"label,omitempty"`
	Created time.Time `json:"created"`
}

type SnapshotsState struct {
	Counter         int              `json:"counter"`
	Snapshots       map[int]Snapshot `json:"snapshots"`
	AppliedSnapshot int              `json:"applied_snapshot,omitempty"`
}

func (ss SnapshotsState) write(p string) error {
	if ss, err := json.Marshal(ss); err != nil {
		return err
	} else {
		if err = ioutil.WriteFile(p, ss, 0600); err != nil {
			return fmt.Errorf("failed to write snapshot state: %v", err)
		}
	}

	return nil
}

func readSnapshotsState(stateFilePath string) (*SnapshotsState, error) {
	b, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		return nil, err
	}

	ss := &SnapshotsState{}

	return ss, json.Unmarshal(b, ss)
}

func NewSnapshot(p, label, comprAlg string, remount bool) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	var (
		diff               = layout.OverlayDiff(p)
		snapshotsDir       = layout.Snapshots(p)
		snapshotsStatePath = layout.SnapshotsState(p)
	)

	ss, err := readSnapshotsState(snapshotsStatePath)
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	ss.Counter++

	snap := Snapshot{
		Id:      ss.Counter,
		Parent:  ss.AppliedSnapshot,
		Label:   label,
		Created: time.Now(),
	}

	snapPath := path.Join(snapshotsDir, layout.SnapshotFilename(snap.Id))
	if err := device.Squash(diff, snapPath, comprAlg); err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}

	if ss.Snapshots == nil {
		ss.Snapshots = make(map[int]Snapshot)
	}

	ss.Snapshots[snap.Id] = snap
	if err := ss.write(snapshotsStatePath); err != nil {
		os.Remove(snapPath)
		return err
	}

	fmt.Println(snap.Id)

	return nil
}

// Only leaves may be deleted, and the leaf must not be AppliedSnapshot
func DeleteSnapshot(p string, snapId int) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	ss, err := readSnapshotsState(layout.SnapshotsState(p))
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	if _, ok := ss.Snapshots[snapId]; !ok {
		return fmt.Errorf("snapshot %d does not exist", snapId)
	}

	if ss.AppliedSnapshot == snapId {
		return fmt.Errorf("snapshot %d is currently being in use", snapId)
	}

	if revDeps := reverseSnapshotDependencies(snapId, ss); revDeps != nil {
		return fmt.Errorf("snapshot %d has dependencies: %v", snapId, revDeps)
	}

	if err := os.Remove(path.Join(layout.Snapshots(p), layout.SnapshotFilename(snapId))); err != nil {
		return fmt.Errorf("failed to remove snapshot: %v", err)
	}

	delete(ss.Snapshots, snapId)
	if err := ss.write(layout.SnapshotsState(p)); err != nil {
		return fmt.Errorf("failed to update snapshots state: %v", err)
	}

	return nil
}

func listHeadLayersForSnapshot(snapId int, ss *SnapshotsState) ([]int, error) {
	if snapId == 0 {
		return nil, nil
	}

	deps, err := snapshotDependencies(snapId, ss)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshot dependencies: %v", err)
	}

	layers := make([]int, len(deps)+1)
	layers[0] = snapId
	copy(layers[1:], deps)

	return layers, nil
}

func ApplySnapshot(p string, snapId int) error {
	// Set up

	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	var (
		head               = layout.Head(p)
		diff               = layout.OverlayDiff(p)
		snapshotsStatePath = layout.SnapshotsState(p)
		snapshotsPath      = layout.Snapshots(p)
		snapshotMountsPath = layout.SnapshotMounts(p)
	)

	ss, err := readSnapshotsState(snapshotsStatePath)
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	if snapId != 0 {
		if _, ok := ss.Snapshots[snapId]; !ok {
			return fmt.Errorf("snapshot %d does not exist", snapId)
		}
	}

	// First, we need to clean up:

	// Unmount overlays

	if err = device.Unmount(p); err != nil {
		return fmt.Errorf("failed to unmount overlay %s: %v", p, err)
	}

	if err = device.Unmount(head); err != nil {
		return fmt.Errorf("failed to unmount HEAD %s: %v", p, err)
	}

	// Unmount all snapshots, if any

	if err := unmountAllSnapshots(snapshotMountsPath); err != nil {
		return err
	}

	// Clean diff

	if err = removeAllIn(diff); err != nil {
		return fmt.Errorf("failed to clean diff: %v", err)
	}

	// Get snapshot dependencies and mount them

	snapLayers, err := listHeadLayersForSnapshot(snapId, ss)
	if err != nil {
		return err
	}

	overlayLayers := make([]string, len(snapLayers)+1)

	for i := range snapLayers {
		mountPoint := path.Join(snapshotMountsPath, layout.SnapshotMountpointTarget(snapLayers[i]))
		overlayLayers[i] = mountPoint

		if err = os.Mkdir(mountPoint, 0700); err != nil {
			return fmt.Errorf("failed to create snapshot mount point %s: %v", mountPoint, err)
		}

		snapshotPath := path.Join(snapshotsPath, layout.SnapshotFilename(snapLayers[i]))
		if err = device.MountSquash(snapshotPath, mountPoint); err != nil {
			return fmt.Errorf("failed to mount snapshot %s: %v", snapshotPath, err)
		}
	}

	overlayLayers[len(snapLayers)] = layout.Orig(p)

	if err = mountHead(head, overlayLayers); err != nil {
		return fmt.Errorf("failed to mount HEAD: %v", err)
	}

	// Mount overlay

	if err = device.OverlayRW(p, diff, layout.OverlayWorkdir(p), head); err != nil {
		return fmt.Errorf("failed to mount overlay: %v", err)
	}

	ss.AppliedSnapshot = snapId
	if err = ss.write(snapshotsStatePath); err != nil {
		return fmt.Errorf("failed to update snapshots state: %v", err)
	}

	return nil
}

func PrintSnapshotsList(p string) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	ss, err := readSnapshotsState(layout.SnapshotsState(p))
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

	fmt.Fprintln(w, "ID\tACTIVE\tCREATED\tLABEL")

	for k, v := range ss.Snapshots {
		active := ' '
		if ss.AppliedSnapshot == k {
			active = '*'
		}
		fmt.Fprintf(w, "%d\t%c\t%s\t%s\n", k, active, v.Created, coalesceStr(v.Label))
	}

	w.Flush()

	return nil
}

func PrintSnapshotDetails(p string, snapId int) error {
	if err := checkTargetAndBaseDirs(p, layout.Base(p)); err != nil {
		return err
	}

	ss, err := readSnapshotsState(layout.SnapshotsState(p))
	if err != nil {
		return fmt.Errorf("failed to read snapshots state: %v", err)
	}

	snap, ok := ss.Snapshots[snapId]
	if !ok {
		return fmt.Errorf("snapshot %d does not exist", snapId)
	}

	deps, err := snapshotDependencies(snapId, ss)
	if err != nil {
		return fmt.Errorf("failed to list snapshot dependencies: %v", deps)
	}

	revDeps := reverseSnapshotDependencies(snapId, ss)

	intSliceToStrSlice := func(xs []int) []string {
		ys := make([]string, len(xs))
		for i, val := range xs {
			ys[i] = strconv.Itoa(val)
		}
		return ys
	}

	depsStr := intSliceToStrSlice(deps)
	revDepsStr := intSliceToStrSlice(revDeps)

	info, err := os.Lstat(path.Join(layout.Snapshots(p), layout.SnapshotFilename(snapId)))
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', 0)

	fmt.Fprintf(w, "Snapshot ID:\t %d\n", snapId)
	fmt.Fprintf(w, "Is active:\t %v\n", ss.AppliedSnapshot == snapId)
	fmt.Fprintf(w, "Created:\t %s\n", snap.Created)
	fmt.Fprintf(w, "Label:\t %s\n", coalesceStr(snap.Label))
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Dependencies:\t %v\n", coalesceStr(strings.Join(depsStr, "->")))
	fmt.Fprintf(w, "Reverse dependencies:\t %v\n", coalesceStr(strings.Join(revDepsStr, ", ")))
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Compressed size:\t %s\n", humanBytes(uint64(info.Size())))

	w.Flush()

	return nil
}

func snapshotDependencies(snapId int, ss *SnapshotsState) ([]int, error) {
	var deps []int
	snapId = ss.Snapshots[snapId].Parent

	for snapId != 0 {
		nextSnapId := ss.Snapshots[snapId].Parent
		if snapId == nextSnapId {
			return nil, fmt.Errorf("circular dependency")
		}

		deps = append(deps, snapId)
		snapId = nextSnapId
	}

	return deps, nil
}

func reverseSnapshotDependencies(snapId int, ss *SnapshotsState) []int {
	var revDeps []int

	if ss.Snapshots != nil {
		for _, val := range ss.Snapshots {
			if val.Parent == snapId {
				revDeps = append(revDeps, val.Id)
			}
		}
	}

	return revDeps
}

func unmountAllSnapshots(snapshotMountsPath string) error {
	iter, err := diriter.NewIter(snapshotMountsPath)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; !iter.AtEnd(); iter.Increment() {
		mountPoint := path.Join(snapshotMountsPath, iter.FileInfo().Name())

		if err = device.UnmountSquash(mountPoint); err != nil {
			return fmt.Errorf("failed to unmount snapshot %s: %v", mountPoint, err)
		}

		if err = os.Remove(mountPoint); err != nil {
			return fmt.Errorf("failed to remove snapshot mount point %s: %v", mountPoint, err)
		}
	}

	return nil
}

func mountHead(into string, lowerDirs []string) error {
	if len(lowerDirs) == 1 {
		return device.BindRO(lowerDirs[0], into)
	}

	return device.OverlayRO(into, lowerDirs...)
}
