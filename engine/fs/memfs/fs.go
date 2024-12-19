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

package memfs

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const filePathSeparator = string(filepath.Separator)
const chmodBits = fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky // Only a subset of bits are allowed to be changed. Documented under os.Chmod()

type FS struct {
	mu   sync.RWMutex
	data map[string]*FileData
	init sync.Once
}

func New() *FS {
	return &FS{}
}

func (m *FS) getData() map[string]*FileData {
	m.init.Do(func() {
		m.data = make(map[string]*FileData)
		// Root should always exist, right?
		// TODO: what about windows?
		root := CreateDir(filePathSeparator)
		SetMode(root, fs.ModeDir|0755)
		m.data[filePathSeparator] = root
	})
	return m.data
}

func (m *FS) Create(name string) (fs.File, error) {
	name = normalizePath(name)
	m.mu.Lock()
	file := CreateFile(name)
	m.getData()[name] = file
	m.registerWithParent(file, 0)
	m.mu.Unlock()
	return NewFileHandle(file), nil
}

func (m *FS) unregisterWithParent(fileName string) error {
	f, err := m.lockfreeOpen(fileName)
	if err != nil {
		return err
	}
	parent := m.findParent(f)
	if parent == nil {
		log.Panic("parent of ", f.Name(), " is nil")
	}

	parent.Lock()
	RemoveFromMemDir(parent, f)
	parent.Unlock()
	return nil
}

func (m *FS) findParent(f *FileData) *FileData {
	pdir, _ := filepath.Split(f.Name())
	pdir = filepath.Clean(pdir)
	pfile, err := m.lockfreeOpen(pdir)
	if err != nil {
		return nil
	}
	return pfile
}

func (m *FS) registerWithParent(f *FileData, perm fs.FileMode) {
	if f == nil {
		return
	}
	parent := m.findParent(f)
	if parent == nil {
		pdir := filepath.Dir(filepath.Clean(f.Name()))
		err := m.lockfreeMkdir(pdir, perm)
		if err != nil {
			//log.Println("Mkdir error:", err)
			return
		}
		parent, err = m.lockfreeOpen(pdir)
		if err != nil {
			//log.Println("Open after Mkdir error:", err)
			return
		}
	}

	parent.Lock()
	InitializeDir(parent)
	AddToMemDir(parent, f)
	parent.Unlock()
}

func (m *FS) lockfreeMkdir(name string, perm fs.FileMode) error {
	name = normalizePath(name)
	x, ok := m.getData()[name]
	if ok {
		// Only return fs.ErrExist if it's a file, not a directory.
		i := FileInfo{FileData: x}
		if !i.IsDir() {
			return fs.ErrExist
		}
	} else {
		item := CreateDir(name)
		SetMode(item, fs.ModeDir|perm)
		m.getData()[name] = item
		m.registerWithParent(item, perm)
	}
	return nil
}

func (m *FS) Mkdir(name string, perm fs.FileMode) error {
	perm &= chmodBits
	name = normalizePath(name)

	m.mu.RLock()
	_, ok := m.getData()[name]
	m.mu.RUnlock()
	if ok {
		return &os.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
	}

	m.mu.Lock()
	item := CreateDir(name)
	SetMode(item, fs.ModeDir|perm)
	m.getData()[name] = item
	m.registerWithParent(item, perm)
	m.mu.Unlock()

	return m.setFileMode(name, perm|fs.ModeDir)
}

func (m *FS) MkdirAll(path string, perm fs.FileMode) error {
	err := m.Mkdir(path, perm)
	if err != nil {
		if err.(*os.PathError).Err == fs.ErrExist {
			return nil
		}
		return err
	}
	return nil
}

// Handle some relative paths
func normalizePath(filePath string) string {
	filePath = path.Clean(filePath)

	switch filePath {
	case ".":
		return filePathSeparator
	case "..":
		return filePathSeparator
	default:
		return filePath
	}
}

func (m *FS) Open(name string) (fs.File, error) {
	f, err := m.open(name)
	if f != nil {
		return NewROFileHandle(f), err
	}
	return nil, err
}

func (m *FS) openWrite(name string) (fs.File, error) {
	f, err := m.open(name)
	if f != nil {
		return NewFileHandle(f), err
	}
	return nil, err
}

func (m *FS) open(name string) (*FileData, error) {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f, nil
}

func (m *FS) lockfreeOpen(name string) (*FileData, error) {
	name = normalizePath(name)
	f, ok := m.getData()[name]
	if ok {
		return f, nil
	} else {
		return nil, fs.ErrNotExist
	}
}

func (m *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	perm &= chmodBits
	chmod := false
	file, err := m.openWrite(name)
	if err == nil && (flag&os.O_EXCL > 0) {
		return nil, &os.PathError{Op: "open", Path: name, Err: fs.ErrExist}
	}
	if os.IsNotExist(err) && (flag&os.O_CREATE > 0) {
		file, err = m.Create(name)
		chmod = true
	}
	if err != nil {
		return nil, err
	}
	if flag == os.O_RDONLY {
		file = NewROFileHandle(file.(*File).Data())
	}
	if flag&os.O_APPEND > 0 {
		fseek, ok := file.(io.Seeker)
		if !ok {
			return nil, fs.ErrPermission
		}
		_, err = fseek.Seek(0, os.SEEK_END)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if flag&os.O_TRUNC > 0 && flag&(os.O_RDWR|os.O_WRONLY) > 0 {
		ftrunc, ok := file.(interface {
			Truncate(size int64) error
		})
		if !ok {
			return nil, fs.ErrPermission
		}
		err = ftrunc.Truncate(0)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if chmod {
		return file, m.setFileMode(name, perm)
	}
	return file, nil
}

func (m *FS) Remove(name string) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.getData()[name]; ok {
		err := m.unregisterWithParent(name)
		if err != nil {
			return &os.PathError{Op: "remove", Path: name, Err: err}
		}
		delete(m.getData(), name)
	} else {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	return nil
}

func (m *FS) RemoveAll(path string) error {
	path = normalizePath(path)
	m.mu.Lock()
	m.unregisterWithParent(path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for p := range m.getData() {
		if strings.HasPrefix(p, path) {
			m.mu.RUnlock()
			m.mu.Lock()
			delete(m.getData(), p)
			m.mu.Unlock()
			m.mu.RLock()
		}
	}
	return nil
}

func (m *FS) Rename(oldname, newname string) error {
	oldname = normalizePath(oldname)
	newname = normalizePath(newname)

	if oldname == newname {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.getData()[oldname]; ok {
		m.mu.RUnlock()
		m.mu.Lock()
		m.unregisterWithParent(oldname)
		fileData := m.getData()[oldname]
		delete(m.getData(), oldname)
		ChangeFileName(fileData, newname)
		m.getData()[newname] = fileData
		m.registerWithParent(fileData, 0)
		m.mu.Unlock()
		m.mu.RLock()
	} else {
		return &os.PathError{Op: "rename", Path: oldname, Err: fs.ErrNotExist}
	}
	return nil
}

func (m *FS) Stat(name string) (fs.FileInfo, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	fi := GetFileInfo(f.(*File).Data())
	return fi, nil
}

func (m *FS) Chmod(name string, mode fs.FileMode) error {
	name = normalizePath(name)
	mode &= chmodBits

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: fs.ErrNotExist}
	}
	prevOtherBits := GetFileInfo(f).Mode() & ^chmodBits

	mode = prevOtherBits | mode
	return m.setFileMode(name, mode)
}

func (m *FS) setFileMode(name string, mode fs.FileMode) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: fs.ErrNotExist}
	}

	m.mu.Lock()
	SetMode(f, mode)
	m.mu.Unlock()

	return nil
}

func (m *FS) Chown(name string, uid, gid int) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chown", Path: name, Err: fs.ErrNotExist}
	}

	SetUID(f, uid)
	SetGID(f, gid)

	return nil
}

func (m *FS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chtimes", Path: name, Err: fs.ErrNotExist}
	}

	m.mu.Lock()
	SetModTime(f, mtime)
	m.mu.Unlock()

	return nil
}

// func (m *FS) List() {
// 	for _, x := range m.data {
// 		y := mem.FileInfo{FileData: x}
// 		fmt.Println(x.Name(), y.Size())
// 	}
// }
