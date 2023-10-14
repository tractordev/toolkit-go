package unionfs

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"time"

	"tractor.dev/toolkit-go/engine/fs/fsutil"
	"tractor.dev/toolkit-go/engine/fs/watchfs"
)

type FS struct {
	base    fs.FS
	overlay fs.FS
}

func New(base, overlay fs.FS) *FS {
	return &FS{
		base:    base,
		overlay: overlay,
	}
}

func isNotExist(err error) bool {
	if e, ok := err.(*os.PathError); ok {
		err = e.Err
	}
	if err == os.ErrNotExist || err == syscall.ENOENT || err == syscall.ENOTDIR {
		return true
	}
	return false
}

func (u *FS) isBaseFile(name string) (bool, error) {
	if _, err := fs.Stat(u.overlay, name); err == nil {
		return false, nil
	}
	_, err := fs.Stat(u.base, name)
	if err != nil {
		if oerr, ok := err.(*os.PathError); ok {
			if oerr.Err == os.ErrNotExist || oerr.Err == syscall.ENOENT || oerr.Err == syscall.ENOTDIR {
				return false, nil
			}
		}
		if err == syscall.ENOENT {
			return false, nil
		}
	}
	return true, err
}

// This function handles the 9 different possibilities caused
// by the union which are the intersection of the following...
//
//	layer: doesn't exist, exists as a file, and exists as a directory
//	base:  doesn't exist, exists as a file, and exists as a directory
func (u *FS) Open(name string) (fs.File, error) {
	// Since the overlay overrides the base we check that first
	b, err := u.isBaseFile(name)
	if err != nil {
		return nil, err
	}

	// If overlay doesn't exist, return the base (base state irrelevant)
	if b {
		return u.base.Open(name)
	}

	// If overlay is a file, return it (base state irrelevant)
	dir, err := fsutil.IsDir(u.overlay, name)
	if err != nil {
		return nil, err
	}
	if !dir {
		return u.overlay.Open(name)
	}

	// Overlay is a directory, base state now matters.
	// Base state has 3 states to check but 2 outcomes:
	// A. It's a file or non-readable in the base (return just the overlay)
	// B. It's an accessible directory in the base (return a UnionFile)

	// If base is file or nonreadable, return overlay
	dir, err = fsutil.IsDir(u.base, name)
	if !dir || err != nil {
		return u.overlay.Open(name)
	}

	// Both base & layer are directories
	// Return union file (if opens are without error)
	bfile, bErr := u.base.Open(name)
	lfile, lErr := u.overlay.Open(name)

	// If either have errors at this point something is very wrong. Return nil and the errors
	if bErr != nil || lErr != nil {
		return nil, fmt.Errorf("BaseErr: %v\nOverlayErr: %v", bErr, lErr)
	}

	return &File{Base: bfile, Layer: lfile}, nil
}

func (u *FS) OpenFile(name string, flag int, perm os.FileMode) (fs.File, error) {
	b, err := u.isBaseFile(name)
	if err != nil {
		return nil, err
	}

	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, fs.ErrPermission
	}
	if b {
		of, ok := u.base.(interface {
			OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
		})
		if !ok {
			return nil, fs.ErrPermission
		}
		return of.OpenFile(name, flag, perm)
	}
	of, ok := u.overlay.(interface {
		OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
	})
	if !ok {
		return nil, fs.ErrPermission
	}
	return of.OpenFile(name, flag, perm)
}

func (u *FS) Stat(name string) (fi fs.FileInfo, err error) {
	fi, err = fs.Stat(u.overlay, name)
	if err != nil {
		if isNotExist(err) {
			return fs.Stat(u.base, name)
		}
		return nil, err
	}
	return fi, nil
}

func (u *FS) Watch(name string, cfg *watchfs.Config) (*watchfs.Watch, error) {
	w, err := watch(u.overlay, name, cfg)
	if err != nil {
		if isNotExist(err) {
			return watch(u.base, name, cfg)
		}
		return nil, err
	}
	return w, nil
}

func (fs *FS) Create(name string) (fs.File, error) {
	return nil, syscall.EPERM
}

func (fs *FS) Mkdir(name string, perm os.FileMode) error {
	return syscall.EPERM
}

func (fs *FS) MkdirAll(path string, perm os.FileMode) error {
	return syscall.EPERM
}

func (fs *FS) Remove(name string) error {
	return syscall.EPERM
}

func (fs *FS) RemoveAll(path string) error {
	return syscall.EPERM
}

func (fs *FS) Rename(oldname, newname string) error {
	return syscall.EPERM
}

func (fs *FS) Chmod(name string, mode os.FileMode) error {
	return syscall.EPERM
}

func (fs *FS) Chown(name string, uid, gid int) error {
	return syscall.EPERM
}

func (fs *FS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return syscall.EPERM
}

type watchFS interface {
	Watch(name string, cfg *watchfs.Config) (*watchfs.Watch, error)
}

func watch(fsys fs.FS, name string, cfg *watchfs.Config) (*watchfs.Watch, error) {
	if fsys, ok := fsys.(watchFS); ok {
		return fsys.Watch(name, cfg)
	}

	return nil, fmt.Errorf("watch %s: operation not supported", name)
}
