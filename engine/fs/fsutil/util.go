package fsutil

import (
	"fmt"
	"io/fs"
	"os"
)

// Remove
// MkDir

func OpenFile(fsys fs.FS, name string, flag int, perm os.FileMode) (fs.File, error) {
	fsopenfile, ok := fsys.(interface {
		OpenFile(name string, flag int, perm os.FileMode) (fs.File, error)
	})
	if !ok {
		return nil, fmt.Errorf("unable to openfile on fs")
	}
	return fsopenfile.OpenFile(name, flag, perm)
}

func MkdirAll(fsys fs.FS, path string, perm fs.FileMode) error {
	fsmkdir, ok := fsys.(interface {
		MkdirAll(path string, perm fs.FileMode) error
	})
	if !ok {
		return fmt.Errorf("unable to mkdirall on fs")
	}
	return fsmkdir.MkdirAll(path, perm)
}

func DirExists(fsys fs.FS, path string) (bool, error) {
	fi, err := fs.Stat(fsys, path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsDir(fsys fs.FS, path string) (bool, error) {
	fi, err := fs.Stat(fsys, path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func IsEmpty(fsys fs.FS, path string) (bool, error) {
	if b, _ := Exists(fsys, path); !b {
		return false, fmt.Errorf("path does not exist: %q", path)
	}
	fi, err := fs.Stat(fsys, path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		f, err := fsys.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()
		list, err := fs.ReadDir(fsys, path)
		return len(list) == 0, nil
	}
	return fi.Size() == 0, nil
}

func Exists(fsys fs.FS, path string) (bool, error) {
	_, err := fs.Stat(fsys, path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
