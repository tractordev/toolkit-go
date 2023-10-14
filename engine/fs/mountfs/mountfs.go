package mountfs

import (
	"io/fs"
	"path/filepath"
	"strings"

	"tractor.dev/toolkit-go/engine/fs/fsutil"
)

// mount.FS assumes the mount takes over at its mount root, ie you cant access
// the base fsys. Using Union() modifier then only does unions on shared
// paths.
func New(fsys fs.FS, root string, mount fs.FS, opts ...Modifier) *FS {
	if root == "" {
		root = "."
	}
	for _, opt := range opts {
		mount = opt(fsys, root, mount)
	}
	return &FS{FS: fsys, mount: mount, root: root}
}

type mountDir struct {
	fs.File
	path    string
	fsys    fs.FS
	dirInfo fs.FileInfo
	dirName string
}

func (f *mountDir) ReadDir(c int) (ofi []fs.DirEntry, err error) {
	d, err := fs.ReadDir(f.fsys, f.path)
	if err != nil {
		return d, err
	}
	d = append(d, f)
	return d, nil
}

func (f *mountDir) Name() string {
	return f.dirName
}
func (f *mountDir) IsDir() bool {
	return true
}
func (f *mountDir) Type() fs.FileMode {
	return f.dirInfo.Mode().Type()
}
func (f *mountDir) Info() (fs.FileInfo, error) {
	return f.dirInfo, nil
}

type FS struct {
	fs.FS

	mount fs.FS
	root  string
}

func (m *FS) Open(name string) (fs.File, error) {
	if m.root == "." {
		return m.mount.Open(name)
	}
	if name == m.root {
		return m.mount.Open(".")
	}
	if name == filepath.Dir(m.root) {
		// TODO: make this work when root is several dirs deep
		fi, err := fs.Stat(m.mount, ".")
		if err != nil {
			return nil, err
		}
		f, err := m.FS.Open(name)
		return &mountDir{File: f, path: name, fsys: m.FS, dirInfo: fi, dirName: filepath.Base(m.root)}, err
	}
	if strings.HasPrefix(name, m.root+"/") {
		return m.mount.Open(strings.TrimPrefix(name, m.root+"/"))
	}
	return m.FS.Open(name)
}

func (m *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	if m.root == "." {
		return fsutil.OpenFile(m.mount, name, flag, perm)
	}
	if name == m.root {
		return fsutil.OpenFile(m.mount, ".", flag, perm)
	}
	if strings.HasPrefix(name, m.root+"/") {
		return fsutil.OpenFile(m.mount, strings.TrimPrefix(name, m.root+"/"), flag, perm)
	}
	return fsutil.OpenFile(m.FS, name, flag, perm)
}

func (m *FS) Stat(name string) (fi fs.FileInfo, err error) {
	if m.root == "." {
		return fs.Stat(m.mount, name)
	}
	if name == m.root {
		return fs.Stat(m.mount, ".")
	}
	if strings.HasPrefix(name, m.root+"/") {
		return fs.Stat(m.mount, strings.TrimPrefix(name, m.root+"/"))
	}
	return fs.Stat(m.FS, name)
}

// func (m *FS) Watch(name string, cfg *watchfs.Config) (*watchfs.Watch, error) {
// 	if m.root == "." {
// 		return Watch(m.mount, name, cfg)
// 	}
// 	if name == m.root {
// 		return Watch(m.mount, ".", cfg)
// 	}
// 	if strings.HasPrefix(name, m.root+"/") {
// 		return Watch(m.mount, strings.TrimPrefix(name, m.root+"/"), cfg)
// 	}
// 	return Watch(m.FS, name, cfg)
// }
