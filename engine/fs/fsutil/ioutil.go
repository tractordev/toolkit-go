package fsutil

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func WriteFile(fsys fs.FS, filename string, data []byte, perm fs.FileMode) error {
	of, ok := fsys.(interface {
		OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)
	})
	if !ok {
		return fs.ErrPermission
	}
	f, err := of.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	fw, ok := f.(io.WriteCloser)
	if !ok {
		f.Close()
		return fs.ErrPermission
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

func TempDir(fsys fs.FS, dir, prefix string) (name string, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		try := filepath.Join(dir, prefix+nextRandom())
		fmkd, ok := fsys.(interface {
			Mkdir(name string, perm fs.FileMode) error
		})
		if !ok {
			return name, fs.ErrPermission
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
