// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workingpathfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FS restricts all operations to a given path within an Fs.
// The given file name to the operations on this Fs will be prepended with
// the base path before calling the base Fs.
// Any file name (after filepath.Clean()) outside this base path will be
// treated as non existing file.
//
// Note that it does not clean the error messages on return, so you may
// reveal the real path on errors.
type FS struct {
	fs.FS
	path string
}

func New(source fs.FS, path string) *FS {
	return &FS{FS: source, path: path}
}

// on a file outside the base path it returns the given file name and an error,
// else the given file with the base path prepended
func (b *FS) RealPath(name string) (path string, err error) {
	if err := validateBasePathName(name); err != nil {
		return name, err
	}

	bpath := filepath.Clean(b.path)
	path = filepath.Clean(filepath.Join(bpath, name))
	if !strings.HasPrefix(path, bpath) {
		return name, os.ErrNotExist
	}

	return path, nil
}

func validateBasePathName(name string) error {
	if runtime.GOOS != "windows" {
		// Not much to do here;
		// the virtual file paths all look absolute on *nix.
		return nil
	}

	// On Windows a common mistake would be to provide an absolute OS path
	// We could strip out the base part, but that would not be very portable.
	if filepath.IsAbs(name) {
		return os.ErrNotExist
	}

	return nil
}

func (b *FS) Chtimes(name string, atime, mtime time.Time) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "chtimes", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Chtimes(name string, atime, mtime time.Time) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Chtimes(name, atime, mtime)
}

func (b *FS) Chmod(name string, mode fs.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "chmod", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Chmod(name string, mode fs.FileMode) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Chmod(name, mode)
}

func (b *FS) Chown(name string, uid, gid int) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "chown", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Chown(name string, uid, gid int) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Chown(name, uid, gid)
}

func (b *FS) Stat(name string) (fi fs.FileInfo, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "stat", Path: name, Err: err}
	}
	return fs.Stat(b.FS, name)
}

func (b *FS) Rename(oldname, newname string) (err error) {
	if oldname, err = b.RealPath(oldname); err != nil {
		return &os.PathError{Op: "rename", Path: oldname, Err: err}
	}
	if newname, err = b.RealPath(newname); err != nil {
		return &os.PathError{Op: "rename", Path: newname, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Rename(oldname, newname string) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Rename(oldname, newname)
}

func (b *FS) RemoveAll(name string) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "remove_all", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		RemoveAll(name string) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.RemoveAll(name)
}

func (b *FS) Remove(name string) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "remove", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Remove(name string) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Remove(name)
}

func (b *FS) OpenFile(name string, flag int, mode fs.FileMode) (f fs.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "openfile", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		OpenFile(name string, flag int, mode fs.FileMode) (f fs.File, err error)
	})
	if !ok {
		return nil, fs.ErrPermission
	}
	srcf, err := fsys.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}
	return srcf, nil
}

func (b *FS) Open(name string) (f fs.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "open", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Open(name string) (f fs.File, err error)
	})
	if !ok {
		return nil, fs.ErrPermission
	}
	srcf, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	return srcf, nil
}

func (b *FS) Mkdir(name string, mode fs.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Mkdir(name string, mode fs.FileMode) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.Mkdir(name, mode)
}

func (b *FS) MkdirAll(name string, mode fs.FileMode) (err error) {
	if name, err = b.RealPath(name); err != nil {
		return &os.PathError{Op: "mkdir", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		MkdirAll(name string, mode fs.FileMode) (err error)
	})
	if !ok {
		return fs.ErrPermission
	}
	return fsys.MkdirAll(name, mode)
}

func (b *FS) Create(name string) (f fs.File, err error) {
	if name, err = b.RealPath(name); err != nil {
		return nil, &os.PathError{Op: "create", Path: name, Err: err}
	}
	fsys, ok := b.FS.(interface {
		Create(name string) (f fs.File, err error)
	})
	if !ok {
		return nil, fs.ErrPermission
	}
	srcf, err := fsys.Create(name)
	if err != nil {
		return nil, err
	}
	return srcf, nil
}
