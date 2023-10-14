package xformfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	vfs "tractor.dev/toolkit-go/engine/fs"
	"tractor.dev/toolkit-go/engine/fs/memfs"
	"tractor.dev/toolkit-go/engine/fs/watchfs"
)

// This package is still experimental and was recently pulled from ui/webiew.go
// so it has some hardcoded bits, but still not 100% on this idea in general

func New(fsys fs.FS) *FS {
	return &FS{FS: fsys}
}

type FS struct {
	fs.FS
	xforms []transform
}

type transform struct {
	suffix string
	fn     func(dst io.Writer, src io.Reader) error
}

func (xfs *FS) Transform(suffix string, fn func(dst io.Writer, src io.Reader) error) {
	xfs.xforms = append(xfs.xforms, transform{
		suffix: suffix,
		fn:     fn,
	})
}

func (xfs *FS) Watch(name string, cfg *watchfs.Config) (*watchfs.Watch, error) {
	wfs, ok := xfs.FS.(vfs.WatchFS)
	if !ok {
		return nil, fmt.Errorf("not supported")
	}
	w, err := wfs.Watch(name, cfg)
	if filepath.Ext(name) == "" && os.IsNotExist(err) {
		exts := []string{".js", ".ts", ".tsx", ".jsx"}
		for _, ext := range exts {
			w, err = xfs.Watch(name+ext, cfg)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return w, err
}

func (xfs *FS) Open(name string) (fs.File, error) {
	f, err := xfs.FS.Open(name)
	if filepath.Ext(name) == "" && os.IsNotExist(err) {
		exts := []string{".js", ".ts", ".tsx", ".jsx"}
		for _, ext := range exts {
			f, err = xfs.Open(name + ext)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return nil, err
	}

	for _, xform := range xfs.xforms {
		if strings.HasSuffix(name, xform.suffix) {
			ff := memfs.NewFileHandle(memfs.CreateFile(name))
			if err := xform.fn(ff, f); err != nil {
				return nil, err
			}
			ff.Seek(0, 0)
			return ff, nil
		}
	}
	return f, nil
}
