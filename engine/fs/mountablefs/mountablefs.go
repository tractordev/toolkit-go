package mountablefs

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
	dir_path = cleanPath(dir_path)

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
	path = cleanPath(path)
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
	for i, m := range host.mounts {
		if strings.HasPrefix(path, m.mountPoint) {
			return true, &host.mounts[i]
		}
	}
	return false, nil
}

func cleanPath(p string) string {
	return filepath.Clean(strings.TrimLeft(p, "/\\"))
}

func trimMountPoint(path string, mntPoint string) string {
	result := strings.TrimPrefix(path, mntPoint)
	result = strings.TrimPrefix(result, string(filepath.Separator))
	
	if result == "" {
		return "."
	} else {
		return result
	}
}

func (host *FS) Chmod(name string, mode fs.FileMode) error  {
	name = cleanPath(name)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	chmodableFS, ok := fsys.(interface {
		Chmod(name string, mode fs.FileMode) error
	})
	if !ok {
		return &fs.PathError{Op: "chmod", Path: name, Err: errors.ErrUnsupported}
	}
	return chmodableFS.Chmod(trimMountPoint(name, prefix), mode)
}

func (host *FS) Chown(name string, uid, gid int) error  {
	name = cleanPath(name)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	chownableFS, ok := fsys.(interface {
		Chown(name string, uid, gid int) error
	})
	if !ok {
		return &fs.PathError{Op: "chown", Path: name, Err: errors.ErrUnsupported}
	}
	return chownableFS.Chown(trimMountPoint(name, prefix), uid, gid)
}

func (host *FS) Chtimes(name string, atime time.Time, mtime time.Time) error  {
	name = cleanPath(name)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	chtimesableFS, ok := fsys.(interface {
		Chtimes(name string, atime time.Time, mtime time.Time) error
	})
	if !ok {
		return &fs.PathError{Op: "chtimes", Path: name, Err: errors.ErrUnsupported}
	}
	return chtimesableFS.Chtimes(trimMountPoint(name, prefix), atime, mtime)
}


func (host *FS) Create(name string) (fs.File, error)  {
	name = cleanPath(name)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	createableFS, ok := fsys.(interface {
		Create(name string) (fs.File, error)
	})
	if !ok {
		return nil, &fs.PathError{Op: "create", Path: name, Err: errors.ErrUnsupported}
	}
	return createableFS.Create(trimMountPoint(name, prefix))
}

func (host *FS) Mkdir(name string, perm fs.FileMode) error  {
	name = cleanPath(name)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(name); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	mkdirableFS, ok := fsys.(interface {
		Mkdir(name string, perm fs.FileMode) error
	})
	if !ok {
		return &fs.PathError{Op: "mkdir", Path: name, Err: errors.ErrUnsupported}
	}
	return mkdirableFS.Mkdir(trimMountPoint(name, prefix), perm)
}

func (host *FS) MkdirAll(path string, perm fs.FileMode) error  {
	path = cleanPath(path)
	var fsys fs.FS
	prefix := ""

	if found, mount := host.isPathInMount(path); found {
		fsys = mount.fsys
		prefix = mount.mountPoint
	} else {
		fsys = host.FS
	}

	mkdirableFS, ok := fsys.(interface {
		MkdirAll(path string, perm fs.FileMode) error
	})
	if !ok {
		return &fs.PathError{Op: "mkdir_all", Path: path, Err: errors.ErrUnsupported}
	}
	return mkdirableFS.MkdirAll(trimMountPoint(path, prefix), perm)
}

func (host *FS) Open(name string) (fs.File, error)  {
	name = cleanPath(name)
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
	name = cleanPath(name)
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
	path = cleanPath(path)
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
	oldname = cleanPath(oldname)
	newname = cleanPath(newname)
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
		return &fs.PathError{Op: "rename", Path: oldname+" -> "+newname, Err: errors.ErrUnsupported}
	}
	return renameableFS.Rename(trimMountPoint(oldname, prefix), trimMountPoint(newname, prefix))
}

// Stat is unecessary since fs.Stat calls Open, which will return a
// File from the correct filesystem anyway. Leaving this here in case 
// it's useful in the future.
// func (host *FS) Stat(name string) (fs.FileInfo, error)  {
//  name = cleanPath(name)
// 	var fsys fs.FS
// 	prefix := ""

// 	if found, mount := host.isPathInMount(name); found {
// 		fsys = mount.fsys
// 		prefix = mount.mountPoint
// 	} else {
// 		fsys = host.FS
// 	}

// 	return fs.Stat(fsys, trimMountPoint(name, prefix))
// }
