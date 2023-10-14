package makefs

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"path/filepath"

	"golang.org/x/text/transform"
	"tractor.dev/toolkit-go/engine/fs/fsutil"
	"tractor.dev/toolkit-go/engine/fs/memfs"
	"tractor.dev/toolkit-go/engine/fs/mountfs"
)

// still wip/experimental

type Opener func(name string) fs.File

func MountOpener(fsys fs.FS, name string, opener Opener) *FS {
	mfs := memfs.New()
	mfs.MkdirAll(filepath.Dir(name), 0755)
	f, err := mfs.Create(name)
	if err != nil {
		panic(err)
	}
	f.Close()
	fsys = mountfs.New(fsys, "", mfs, mountfs.Union())
	return New(fsys, name, opener)
}

func New(fsys fs.FS, pattern string, opener Opener) *FS {
	return &FS{FS: fsys, pattern: pattern, opener: opener}
}

type FS struct {
	fs.FS

	opener  Opener
	pattern string
}

func (m *FS) Open(name string) (fs.File, error) {
	if ok, err := filepath.Match(m.pattern, name); ok && err == nil {
		return m.opener(name), nil
	}
	return m.FS.Open(name)
}

func (m *FS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	if ok, err := filepath.Match(m.pattern, name); ok && err == nil {
		return m.opener(name), nil
	}
	return fsutil.OpenFile(m.FS, name, flag, perm)
}

func (m *FS) Stat(name string) (fi fs.FileInfo, err error) {
	if ok, err := filepath.Match(m.pattern, name); ok && err == nil {
		return m.opener(name).Stat()
	}
	return fs.Stat(m.FS, name)
}

type OpenFile struct {
	Path string
	File fs.File
}

// MakeFunc takes open Files and does some operation returning bytes.
type MakeFunc func(files []OpenFile) ([]byte, error)

// MakeOpener returns an Opener that performs a MakeFunc using optional dependency filepaths.
// The idea is to replicate Make task semantics.
func MakeOpener(fsys fs.FS, fn MakeFunc, deps ...string) Opener {
	return func(filename string) fs.File {
		var files []OpenFile
		if len(deps) > 0 {
			fs.WalkDir(fsys, "", func(path string, info fs.DirEntry, err error) error {
				if info.IsDir() {
					return nil
				}
				for _, dep := range deps {
					ok, err := filepath.Match(dep, path)
					if !ok || err != nil {
						continue
					}
					f, err := fsys.Open(path)
					if err != nil {
						log.Println(err)
						continue
					}
					files = append(files, OpenFile{Path: path, File: f})
				}
				return nil
			})
		}

		f := memfs.NewFileHandle(memfs.CreateFile(filepath.Base(filename)))
		defer f.Seek(0, 0)

		b, err := fn(files)
		if err != nil {
			log.Println(err)
			f.Write([]byte(err.Error()))
			return f
		}

		_, err = f.Write(b)
		if err != nil {
			log.Println(err)
			f.Write([]byte(err.Error()))
		}
		return f
	}
}

// TransformFrom returns an Opener that performs a transform on the given source file.
// It is a more specific API for MakeOpener focusing on a single file and the transform.Transformer interface.
func TransformFrom(fsys fs.FS, name string, xform transform.Transformer) Opener {
	return MakeOpener(fsys, func(files []OpenFile) ([]byte, error) {
		b, err := io.ReadAll(files[0].File)
		if err != nil {
			return []byte{}, err
		}
		var src io.Reader
		src = bytes.NewBuffer(b)
		if xform != nil {
			src = transform.NewReader(src, xform)
		}
		var dst bytes.Buffer
		if _, err := io.Copy(&dst, src); err != nil {
			return []byte{}, err
		}
		return dst.Bytes(), nil
	}, name)
}
