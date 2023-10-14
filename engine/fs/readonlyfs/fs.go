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

package readonlyfs

import (
	"io/fs"
	"os"
	"syscall"
	"time"
)

type FS struct {
	fs.FS
}

func New(fsys fs.FS) *FS {
	return &FS{FS: fsys}
}

func (r *FS) Chtimes(n string, a, m time.Time) error {
	return fs.ErrPermission
}

func (r *FS) Chmod(n string, m fs.FileMode) error {
	return fs.ErrPermission
}

func (r *FS) Chown(n string, uid, gid int) error {
	return fs.ErrPermission
}

func (r *FS) Rename(o, n string) error {
	return fs.ErrPermission
}

func (r *FS) RemoveAll(p string) error {
	return fs.ErrPermission
}

func (r *FS) Remove(n string) error {
	return fs.ErrPermission
}

func (r *FS) Mkdir(n string, p fs.FileMode) error {
	return fs.ErrPermission
}

func (r *FS) MkdirAll(n string, p fs.FileMode) error {
	return fs.ErrPermission
}

func (r *FS) Create(n string) (fs.File, error) {
	return nil, fs.ErrPermission
}

func (r *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	if flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, fs.ErrPermission
	}
	of, ok := r.FS.(interface {
		OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
	})
	if !ok {
		return nil, fs.ErrPermission
	}
	return of.OpenFile(name, flag, perm)
}
