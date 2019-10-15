eph (ephemeral) is a simple ramdisk management tool for Linux

#### What's it good for

Provided you have enough spare, disposable RAM, you can:
* create throw-away scratch space (e.g. artefacts from builds, cache, temporary files)
* perform high IOPs tasks in a buffer before the data is written back to your persistent storage
* overlay existing data with a ramdisk
* overlay read-only directories with a writable ramdisk
* create intermediate "check-points" between writes using snapshots

Most of the items mentioned above could be achieved manually with just tmpfs, overlayfs and squashfs file-systems. eph automates and streamlines this process.

## Features

* creating blank tmpfs ramdisks
* creating ramdisks over existing directories
* displaying differences between on-disk data and a ramdisk
* commiting changes from a ramdisk to persistent storage
* online snapshotting; applying (recovering from) snapshots is done offline, as it requires a remount
* snapshot compression via squashfs

## Building from source, installation

If you don't want to build from source, you may download a pre-built executable from the [releases](https://github.com/gman0/eph/releases) page. Otherwise keep on reading.

### Dependencies

Snapshotting requires `squashfs-tools` to be installed on your system and accessible from the PATH environment variable.

### Building from source

eph is written in Go, you'll need the Go toolchain 1.12+ to build eph:
```bash
# Debian-based distributions
sudo apt install go
# Arch-based distributions
sudo pacman -S go
# Snap package
sudo snap install --classic --channel=1.12/stable go
```

Clone from git:
```bash
git clone https://github.com/gman0/eph.git
```

Build:
```bash
cd eph
make
```

This will compile eph into a single executable file in `_output/eph`.

### Installation

Run:
```bash
sudo make install
```

By default, eph is installed under `/usr/local`. Set `PREFIX` variable to change the install location.

## Using eph, examples

Please keep in mind that its RAM's nature to be ephemeral and it's impossible to recover the data off the ramdisk after it's been unmounted, or the computer has been restarted.

Note that in order to allow eph to modify mounts, it needs to be run as root or an equivalent user with `SYS_ADMIN` capabilities.

Each command listed below has a `--help` flag that goes into more detail.

**Creating a blank ramdisk**

```bash
sudo eph create /home/foo/new --quota 1G
```

`create` command creates a new ramdisk in the specified target location. `--quota` is an optional flag to set the quota on the ramdisk. Note that the quota represents only the upper limit on the data volume, the ramdisk will only occupy the actual amount of space stored in it.

You may find one extra directory created in `/home/foo/.eph.new` -- eph root, where the ramdisk's internals are stored. It's not advised to tamper with its contents.

Internally, eph uses tmpfs to create and handle ramdisks, which means the data may still be swapped out of RAM and written onto disk if swapping is enabled. See [tmpfs docs](https://www.kernel.org/doc/Documentation/filesystems/tmpfs.txt) for more info.

**Overlaying an existing directory**

```bash
sudo eph create /home/foo/bar --overlay
```

`create --overlay` creates a new ramdisk and overlays it over an existing directory. The original data is accessed as read-only and any modifications (creating/moving/deleting files and directories) to the target directory are stored inside the ramdisk. eph uses OverlayFS for file-system union (see [overlayfs docs](https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt) for more info).

Differences between the ramdisk and original on-disk data are stored in per-file granularity as opposed to per-block, which means that if a file that exists on disk is being modified, it must be internally copied over to the ramdisk first. All reads and writes to that file then continue normally to its ramdisk copy. Keep this in mind in regards to the quota settings.

**Displaying differences**

```bash
sudo eph status /home/foo/bar
```

`status` command displays differences between ramdisk and original on-disk data. Its output may look like this:
```
M modified-file
A new-file
D dont-need-this-file
a new-dir
```

Each file or directory name is prefixed with a status code `A`, `M` or `D`, which stand for _added_, _modified_ or _deleted_ respectively. Status codes for directories are in lower-case, all non-directories use upper-case codes. All paths in the output are relative to the path specified for the `status` command.

**Discarding a ramdisk**

```bash
sudo eph discard /home/foo/bar
```

`discard` command unmounts the ramdisk, *irreversably* discarding any changes made to the original data.

**Writing changes from ramdisk to persistent storage**

```bash
sudo eph merge /home/foo/bar
```

`merge` commits all data from the ramdisk to its target location and unmounts the ramdisk. Outputs a list of changes similar to the `status` command.

**Managing ramdisk snapshots**

Snapshotting requires `squashfs-tools` to be installed on your system and accessible from the PATH environment variable.

```bash
sudo eph snapshot new /home/foo/bar
```

`snapshot new` command creates a new snapshot of the ramdisk. Snapshots are created online (i.e. no remounts are needed, inodes are preserved) but users must make sure no writes occur while the command is running, otherwise contents of the snapshot may be corrupted. The command outputs the snapshot ID used for identification of the snapshot. The ID is a numerical value, increasing monotonically from 1.

Note that eph stores the snapshots inside the ramdisk, which means they contribute to overall ramdisk space consumption.

```bash
sudo eph snapshot apply /home/foo/bar --id 1
```

`snapshot apply --id` restores the ramdisk to the state when the snapshot was taken. Snapshot ID `0` is reserved ID with special meaning: applying to ID 0 resets the ramdisk to the original data.

You can also use `snapshot new --apply` to create a new snapshot and apply it immediately.

A combination of successive `snapshot new` and `snapshot apply` commands makes it possible to stack snapshots, e.g. applying snapshot ID 1 and creating another snapshot with ID 2 means that snapshot 2 depends on snapshot 1, effectively creating tree structures (i.e branching off of snapshots).

Note that applying a snapshot is done offline: all files and directories in the target location must be closed before executing `snapshot apply`, all inodes will be invalidated.

```bash
sudo eph snapshot list /home/foo/bar
```

`snapshot list` lists all snapshots of a ramdisk.

```bash
sudo eph snapshot show /home/foo/bar --id 1
```

`snapshot show --id` shows details for a ramdisk snapshot.

```bash
sudo eph snapshot delete /home/foo/bar --id 1
```

`snapshot delete --id` deletes a ramdisk snapshot. It must not have any child snapshots and must not be currently active.

**Setting ramdisk quota**

```bash
sudo eph set-quota /home/foo/bar --quota 5G
```

`set-quota --quota` sets volume quota of a ramdisk to the specified size.

## Troubleshooting

If an error occurs, you may always find your original data in the `orig` directory in eph root (e.g. `/home/foo/.eph.bar/orig` for `/home/foo/bar` target location). Restoring from errors during `merge` is currently problematic as it may leave the original data in an inconsistent state (because only _some_ files were copied over from the ramdisk) - this may be improved in future versions.
