package mountablefs

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"syscall"

	"tractor.dev/toolkit-go/engine/fs/fsutil"
)

type mountedFSDir struct {
	fsys fs.FS
	mountPoint string
}

type FS struct {
	fs.FS
	mounts []mountedFSDir
}

func New(fsys fs.FS) *FS {
	return &FS{FS: fsys, mounts: make([]mountedFSDir, 0, 1)}
}

func (host *FS) Mount(fsys fs.FS, dir_path string) error {
	dir_path = filepath.Clean(dir_path)

	fi, err := fs.Stat(host, dir_path)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return &fs.PathError{Op: "mount", Path: dir_path, Err: fs.ErrInvalid}
	}
	if found, _ := host.isPathInMount(dir_path); found {
		return &fs.PathError{Op: "mount", Path: dir_path, Err: fs.ErrExist}
	}

	host.mounts = append(host.mounts, mountedFSDir{fsys: fsys, mountPoint: dir_path})
	return nil
}

func (host *FS) Unmount(path string) error {
	path = filepath.Clean(path)
	for i, m := range host.mounts {
		if path == m.mountPoint {
			host.mounts = remove(host.mounts, i)
			return nil
		}
	}

	return &fs.PathError{Op: "unmount", Path: path, Err: fs.ErrInvalid}
}

func remove(s []mountedFSDir, i int) []mountedFSDir {
    s[i] = s[len(s)-1]
    return s[:len(s)-1]
}

func (host *FS) isPathInMount(path string) (bool, *mountedFSDir) {
	path = filepath.Clean(path)
	for i, m := range host.mounts {
		if strings.HasPrefix(path, m.mountPoint) {
			return true, &host.mounts[i]
		}
	}
	return false, nil
}

func trimMountPoint(path string, mntPoint string) string {
	// Clean replaces path seperators with the native one, which
	// requires us to do extra cleaning when handling mountPoint paths.
	result := strings.TrimPrefix(filepath.Clean(path), mntPoint)
	return filepath.Clean(strings.TrimPrefix(result, string(filepath.Separator)))
}

// TODO:
// func (host *FS) Chmod(name string, mode fs.FileMode) error  {}
// func (host *FS) Chown(name string, uid, gid int) error  {}
// func (host *FS) Chtimes(name string, atime time.Time, mtime time.Time) error  {}
// func (host *FS) Create(name string) (fs.File, error)  {}
// func (host *FS) Mkdir(name string, perm fs.FileMode) error  {}
// func (host *FS) MkdirAll(path string, perm fs.FileMode) error  {}

func (host *FS) Open(name string) (fs.File, error)  {
	if !fs.ValidPath(name) { // TODO: may be redundant
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if found, mount := host.isPathInMount(name); found {
		return mount.fsys.Open(trimMountPoint(name, mount.mountPoint))
	}

	return host.FS.Open(name)
}

func (host *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)  {
	if found, mount := host.isPathInMount(name); found {
		return fsutil.OpenFile(mount.fsys, trimMountPoint(name, mount.mountPoint), flag, perm)
	} else {
		return fsutil.OpenFile(host.FS, name, flag, perm)
	}
}

func (host *FS) Remove(name string) error  {
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		// TODO: error if trying to remove mountPoint?
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	removableFS, ok := fsys.(interface {
		Remove(name string) error
	})
	if !ok {
		return fmt.Errorf("remove: %w", errors.ErrUnsupported)
	}
	return removableFS.Remove(trimMountPoint(name, prefix))
}

func (host *FS) RemoveAll(path string) error  {
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(path); found {
		fsys = mount.fsys
		// TODO: error if trying to remove mountPoint?
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	removableFS, ok := fsys.(interface {
		RemoveAll(path string) error
	})
	if !ok {
		// TODO: default implementation which depends on fsys supporting Remove
		return fmt.Errorf("remove_all: %w", errors.ErrUnsupported)
	}
	return removableFS.RemoveAll(trimMountPoint(path, prefix))
}

func (host *FS) Rename(oldname, newname string) error  {
	var fsys fs.FS
	prefix := ""

	// error if both paths aren't in the same filesystem
	if found, oldMount := host.isPathInMount(oldname); found {
		if found, newMount := host.isPathInMount(newname); found {
			if oldMount != newMount {
				return &fs.PathError{Op: "rename", Path: oldname+" -> "+newname, Err: syscall.EXDEV}		
			}

			// TODO: error if trying to rename mountPoint?
			fsys = newMount.fsys
			prefix = newMount.mountPoint
		} else {
			return &fs.PathError{Op: "rename", Path: oldname+" -> "+newname, Err: syscall.EXDEV}		
		}	
	} else {
		if found, _ := host.isPathInMount(newname); found {
			return &fs.PathError{Op: "rename", Path: oldname+" -> "+newname, Err: syscall.EXDEV}		
		}

		fsys = host.FS
	}

	renameableFS, ok := fsys.(interface {
		Rename(oldname, newname string) error
	})
	if !ok {
		return fmt.Errorf("rename: %w", errors.ErrUnsupported)
	}
	return renameableFS.Rename(trimMountPoint(oldname, prefix), trimMountPoint(newname, prefix))
}

func (host *FS) Stat(name string) (fs.FileInfo, error)  {
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	return fs.Stat(fsys, trimMountPoint(name, prefix))
}
