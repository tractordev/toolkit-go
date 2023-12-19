package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// todo: Remove
// todo: MkDir

func OpenFile(fsys FS, name string, flag int, perm os.FileMode) (File, error) {
	fsopenfile, ok := fsys.(interface {
		OpenFile(name string, flag int, perm os.FileMode) (File, error)
	})
	if !ok {
		return nil, fmt.Errorf("unable to openfile on fs")
	}
	return fsopenfile.OpenFile(name, flag, perm)
}

func MkdirAll(fsys FS, path string, perm FileMode) error {
	fsmkdir, ok := fsys.(interface {
		MkdirAll(path string, perm FileMode) error
	})
	if !ok {
		return fmt.Errorf("unable to mkdirall on fs")
	}
	return fsmkdir.MkdirAll(path, perm)
}

func DirExists(fsys FS, path string) (bool, error) {
	fi, err := Stat(fsys, path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func IsDir(fsys FS, path string) (bool, error) {
	fi, err := Stat(fsys, path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func IsEmpty(fsys FS, path string) (bool, error) {
	if b, _ := Exists(fsys, path); !b {
		return false, fmt.Errorf("path does not exist: %q", path)
	}
	fi, err := Stat(fsys, path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		f, err := fsys.Open(path)
		if err != nil {
			return false, err
		}
		defer f.Close()
		list, err := ReadDir(fsys, path)
		return len(list) == 0, nil
	}
	return fi.Size() == 0, nil
}

func Exists(fsys FS, path string) (bool, error) {
	_, err := Stat(fsys, path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func WriteFile(fsys FS, filename string, data []byte, perm FileMode) error {
	of, ok := fsys.(interface {
		OpenFile(name string, flag int, perm FileMode) (File, error)
	})
	if !ok {
		return ErrPermission
	}
	f, err := of.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	fw, ok := f.(io.WriteCloser)
	if !ok {
		f.Close()
		return ErrPermission
	}
	n, err := fw.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := fw.Close(); err == nil {
		err = err1
	}
	return err
}

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextRandom() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func TempDir(fsys FS, dir, prefix string) (name string, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		try := filepath.Join(dir, prefix+nextRandom())
		fmkd, ok := fsys.(interface {
			Mkdir(name string, perm FileMode) error
		})
		if !ok {
			return name, ErrPermission
		}
		err = fmkd.Mkdir(try, 0700)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				randmu.Lock()
				rand = reseed()
				randmu.Unlock()
			}
			continue
		}
		if err == nil {
			name = try
		}
		break
	}
	return
}
