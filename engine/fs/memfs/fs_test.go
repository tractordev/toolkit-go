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
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"tractor.dev/toolkit-go/engine/fs/fsutil"
)

var Fss = []*FS{{}}

var testRegistry map[*FS][]string = make(map[*FS][]string)

func removeAllTestFiles(t *testing.T) {
	for fs, list := range testRegistry {
		for _, path := range list {
			if err := fs.RemoveAll(path); err != nil {
				t.Error(err)
			}
		}
	}
	testRegistry = make(map[*FS][]string)
}

func testDir(fsys *FS) string {
	name, err := fsutil.TempDir(fsys, "", "vfs")
	if err != nil {
		panic(fmt.Sprint("unable to work with test dir", err))
	}
	testRegistry[fsys] = append(testRegistry[fsys], name)

	return name
}

func fsName(fs *FS) string {
	return filepath.Base(reflect.TypeOf(fs).Elem().PkgPath()) + ".FS"
}

func TestNormalizePath(t *testing.T) {
	type test struct {
		input    string
		expected string
	}

	data := []test{
		{".", filePathSeparator},
		{"./", filePathSeparator},
		{"..", filePathSeparator},
		{"../", filePathSeparator},
		{"./..", filePathSeparator},
		{"./../", filePathSeparator},
	}

	for i, d := range data {
		cpath := normalizePath(d.input)
		if d.expected != cpath {
			t.Errorf("Test %d failed. Expected %q got %q", i, d.expected, cpath)
		}
	}
}

func TestPathErrors(t *testing.T) {
	path := filepath.Join(".", "some", "path")
	path2 := filepath.Join(".", "different", "path")
	fs := New()
	perm := os.FileMode(0755)
	uid := 1000
	gid := 1000

	// relevant functions:
	// func (m *MemMapFs) Chmod(name string, mode os.FileMode) error
	// func (m *MemMapFs) Chtimes(name string, atime time.Time, mtime time.Time) error
	// func (m *MemMapFs) Create(name string) (File, error)
	// func (m *MemMapFs) Mkdir(name string, perm os.FileMode) error
	// func (m *MemMapFs) MkdirAll(path string, perm os.FileMode) error
	// func (m *MemMapFs) Open(name string) (File, error)
	// func (m *MemMapFs) OpenFile(name string, flag int, perm os.FileMode) (File, error)
	// func (m *MemMapFs) Remove(name string) error
	// func (m *MemMapFs) Rename(oldname, newname string) error
	// func (m *MemMapFs) Stat(name string) (os.FileInfo, error)

	err := fs.Chmod(path, perm)
	checkPathError(t, err, "Chmod")

	err = fs.Chown(path, uid, gid)
	checkPathError(t, err, "Chown")

	err = fs.Chtimes(path, time.Now(), time.Now())
	checkPathError(t, err, "Chtimes")

	// fs.Create doesn't return an error

	err = fs.Mkdir(path2, perm)
	if err != nil {
		t.Error(err)
	}
	err = fs.Mkdir(path2, perm)
	checkPathError(t, err, "Mkdir")

	err = fs.MkdirAll(path2, perm)
	if err != nil {
		t.Error("MkdirAll:", err)
	}

	_, err = fs.Open(path)
	checkPathError(t, err, "Open")

	_, err = fs.OpenFile(path, os.O_RDWR, perm)
	checkPathError(t, err, "OpenFile")

	err = fs.Remove(path)
	checkPathError(t, err, "Remove")

	err = fs.RemoveAll(path)
	if err != nil {
		t.Error("RemoveAll:", err)
	}

	err = fs.Rename(path, path2)
	checkPathError(t, err, "Rename")

	_, err = fs.Stat(path)
	checkPathError(t, err, "Stat")
}

func checkPathError(t *testing.T, err error, op string) {
	pathErr, ok := err.(*os.PathError)
	if !ok {
		t.Error(op+":", err, "is not a os.PathError")
		return
	}
	_, ok = pathErr.Err.(*os.PathError)
	if ok {
		t.Error(op+":", err, "contains another os.PathError")
	}
}

// Ensure os.O_EXCL is correctly handled.
func TestOpenFileExcl(t *testing.T) {
	const fileName = "/myFileTest"
	const fileMode = os.FileMode(0765)

	fs := New()

	// First creation should succeed.
	f, err := fs.OpenFile(fileName, os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		t.Errorf("OpenFile Create Excl failed: %s", err)
		return
	}
	f.Close()

	// Second creation should fail.
	_, err = fs.OpenFile(fileName, os.O_CREATE|os.O_EXCL, fileMode)
	if err == nil {
		t.Errorf("OpenFile Create Excl should have failed, but it didn't")
	}
	checkPathError(t, err, "Open")
}

// Ensure Permissions are set on OpenFile/Mkdir/MkdirAll
func TestPermSet(t *testing.T) {
	const fileName = "/myFileTest"
	const dirPath = "/myDirTest"
	const dirPathAll = "/my/path/to/dir"

	const fileMode = os.FileMode(0765)
	// directories will also have the directory bit set
	const dirMode = fileMode | os.ModeDir

	fs := New()

	// Test Openfile
	f, err := fs.OpenFile(fileName, os.O_CREATE, fileMode)
	if err != nil {
		t.Errorf("OpenFile Create failed: %s", err)
		return
	}
	f.Close()

	s, err := fs.Stat(fileName)
	if err != nil {
		t.Errorf("Stat failed: %s", err)
		return
	}
	if s.Mode().String() != fileMode.String() {
		t.Errorf("Permissions Incorrect: %s != %s", s.Mode().String(), fileMode.String())
		return
	}

	// Test Mkdir
	err = fs.Mkdir(dirPath, dirMode)
	if err != nil {
		t.Errorf("MkDir Create failed: %s", err)
		return
	}
	s, err = fs.Stat(dirPath)
	if err != nil {
		t.Errorf("Stat failed: %s", err)
		return
	}
	// sets File
	if s.Mode().String() != dirMode.String() {
		t.Errorf("Permissions Incorrect: %s != %s", s.Mode().String(), dirMode.String())
		return
	}

	// Test MkdirAll
	err = fs.MkdirAll(dirPathAll, dirMode)
	if err != nil {
		t.Errorf("MkDir Create failed: %s", err)
		return
	}
	s, err = fs.Stat(dirPathAll)
	if err != nil {
		t.Errorf("Stat failed: %s", err)
		return
	}
	if s.Mode().String() != dirMode.String() {
		t.Errorf("Permissions Incorrect: %s != %s", s.Mode().String(), dirMode.String())
		return
	}
}

// Fails if multiple file objects use the same file.at counter in MemMapFs
func TestMultipleOpenFiles(t *testing.T) {
	defer removeAllTestFiles(t)
	const fileName = "afero-demo2.txt"

	var data = make([][]byte, len(Fss))

	for i, fsys := range Fss {
		dir := testDir(fsys)
		path := filepath.Join(dir, fileName)
		fh1, err := fsys.Create(path)
		if err != nil {
			t.Error("fs.Create failed: " + err.Error())
		}
		fw1, ok := fh1.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw1.Write([]byte("test"))
		if err != nil {
			t.Error("fh.Write failed: " + err.Error())
		}
		fs1, ok := fh1.(io.Seeker)
		if !ok {
			t.Fatal("file not a io.Seeker")
		}
		_, err = fs1.Seek(0, os.SEEK_SET)
		if err != nil {
			t.Error(err)
		}

		fh2, err := fsys.OpenFile(path, os.O_RDWR, 0777)
		if err != nil {
			t.Error("fs.OpenFile failed: " + err.Error())
		}
		fs2, ok := fh2.(io.Seeker)
		if !ok {
			t.Fatal("file not a io.Seeker")
		}
		_, err = fs2.Seek(0, os.SEEK_END)
		if err != nil {
			t.Error(err)
		}
		fw2, ok := fh2.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw2.Write([]byte("data"))
		if err != nil {
			t.Error(err)
		}
		err = fh2.Close()
		if err != nil {
			t.Error(err)
		}

		_, err = fw1.Write([]byte("data"))
		if err != nil {
			t.Error(err)
		}
		err = fh1.Close()
		if err != nil {
			t.Error(err)
		}
		// the file now should contain "datadata"
		data[i], err = fs.ReadFile(fsys, path)
		if err != nil {
			t.Error(err)
		}
	}

	for i, fs := range Fss {
		if i == 0 {
			continue
		}
		if string(data[0]) != string(data[i]) {
			t.Errorf("%s and %s don't behave the same\n"+
				"%s: \"%s\"\n%s: \"%s\"\n",
				fsName(Fss[0]), fsName(fs), fsName(Fss[0]), data[0], fsName(fs), data[i])
		}
	}
}

// Test if file.Write() fails when opened as read only
func TestReadOnly(t *testing.T) {
	defer removeAllTestFiles(t)
	const fileName = "afero-demo.txt"

	for _, fs := range Fss {
		dir := testDir(fs)
		path := filepath.Join(dir, fileName)

		f, err := fs.Create(path)
		if err != nil {
			t.Error(fsName(fs)+":", "fs.Create failed: "+err.Error())
		}
		fw, ok := f.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw.Write([]byte("test"))
		if err != nil {
			t.Error(fsName(fs)+":", "Write failed: "+err.Error())
		}
		f.Close()

		f, err = fs.Open(path)
		if err != nil {
			t.Error("fs.Open failed: " + err.Error())
		}
		fw, ok = f.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw.Write([]byte("data"))
		if err == nil {
			t.Error(fsName(fs)+":", "No write error")
		}
		f.Close()

		f, err = fs.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			t.Error("fs.Open failed: " + err.Error())
		}
		fw, ok = f.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw.Write([]byte("data"))
		if err == nil {
			t.Error(fsName(fs)+":", "No write error")
		}
		f.Close()
	}
}

func TestWriteCloseTime(t *testing.T) {
	defer removeAllTestFiles(t)
	const fileName = "afero-demo.txt"

	for _, fs := range Fss {
		dir := testDir(fs)
		path := filepath.Join(dir, fileName)

		f, err := fs.Create(path)
		if err != nil {
			t.Error(fsName(fs)+":", "fs.Create failed: "+err.Error())
		}
		f.Close()

		f, err = fs.Create(path)
		if err != nil {
			t.Error(fsName(fs)+":", "fs.Create failed: "+err.Error())
		}
		fi, err := f.Stat()
		if err != nil {
			t.Error(fsName(fs)+":", "Stat failed: "+err.Error())
		}
		timeBefore := fi.ModTime()

		// sorry for the delay, but we have to make sure time advances,
		// also on non Un*x systems...
		switch runtime.GOOS {
		case "windows":
			time.Sleep(2 * time.Second)
		case "darwin":
			time.Sleep(1 * time.Second)
		default: // depending on the FS, this may work with < 1 second, on my old ext3 it does not
			time.Sleep(1 * time.Second)
		}

		fw, ok := f.(io.Writer)
		if !ok {
			t.Fatal("file not a io.Writer")
		}
		_, err = fw.Write([]byte("test"))
		if err != nil {
			t.Error(fsName(fs)+":", "Write failed: "+err.Error())
		}
		f.Close()
		fi, err = fs.Stat(path)
		if err != nil {
			t.Error(fsName(fs)+":", "fs.Stat failed: "+err.Error())
		}
		if fi.ModTime().Equal(timeBefore) {
			t.Error(fsName(fs)+":", "ModTime was not set on Close()")
		}
	}
}

// This test should be run with the race detector on:
// go test -race -v -timeout 10s -run TestRacingDeleteAndClose
func TestRacingDeleteAndClose(t *testing.T) {
	fs := New()
	pathname := "testfile"
	f, err := fs.Create(pathname)
	if err != nil {
		t.Fatal(err)
	}

	in := make(chan bool)

	go func() {
		<-in
		f.Close()
	}()
	go func() {
		<-in
		fs.Remove(pathname)
	}()
	close(in)
}

// This test should be run with the race detector on:
// go test -run TestMemFsDataRace -race
func TestMemFsDataRace(t *testing.T) {
	const dir = "test_dir"
	fsys := New()

	if err := fsys.MkdirAll(dir, 0777); err != nil {
		t.Fatal(err)
	}

	const n = 1000
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			fname := filepath.Join(dir, fmt.Sprintf("%d.txt", i))
			if err := fsutil.WriteFile(fsys, fname, []byte(""), 0777); err != nil {
				panic(err)
			}
			if err := fsys.Remove(fname); err != nil {
				panic(err)
			}
		}
	}()

loop:
	for {
		select {
		case <-done:
			break loop
		default:
			_, err := fs.ReadDir(fsys, dir)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

// root is a directory
func TestMemFsRootDirMode(t *testing.T) {
	t.Parallel()

	fs := New()
	info, err := fs.Stat("/")
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
	if !info.Mode().IsDir() {
		t.Errorf("FileMode is not directory, is %s", info.Mode().String())
	}
}

// MkdirAll creates intermediate directories with correct mode
func TestMemFsMkdirAllMode(t *testing.T) {
	t.Parallel()

	fs := New()
	err := fs.MkdirAll("/a/b/c", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := fs.Stat("/a")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsDir() {
		t.Error("/a: mode is not directory")
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Errorf("/a: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
	info, err = fs.Stat("/a/b")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsDir() {
		t.Error("/a/b: mode is not directory")
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Errorf("/a/b: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
	info, err = fs.Stat("/a/b/c")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsDir() {
		t.Error("/a/b/c: mode is not directory")
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Errorf("/a/b/c: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
}

// MkdirAll does not change permissions of already-existing directories
func TestMemFsMkdirAllNoClobber(t *testing.T) {
	t.Parallel()

	fs := New()
	err := fs.MkdirAll("/a/b/c", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := fs.Stat("/a/b")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Errorf("/a/b: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
	err = fs.MkdirAll("/a/b/c/d/e/f", 0710)
	// '/a/b' is unchanged
	if err != nil {
		t.Fatal(err)
	}
	info, err = fs.Stat("/a/b")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Errorf("/a/b: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
	// new directories created with proper permissions
	info, err = fs.Stat("/a/b/c/d")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0710) {
		t.Errorf("/a/b/c/d: wrong permissions, expected drwx--x---, got %s", info.Mode())
	}
	info, err = fs.Stat("/a/b/c/d/e")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0710) {
		t.Errorf("/a/b/c/d/e: wrong permissions, expected drwx--x---, got %s", info.Mode())
	}
	info, err = fs.Stat("/a/b/c/d/e/f")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0710) {
		t.Errorf("/a/b/c/d/e/f: wrong permissions, expected drwx--x---, got %s", info.Mode())
	}
}

func TestMemFsDirMode(t *testing.T) {
	fs := New()
	err := fs.Mkdir("/testDir1", 0644)
	if err != nil {
		t.Error(err)
	}
	err = fs.MkdirAll("/sub/testDir2", 0644)
	if err != nil {
		t.Error(err)
	}
	info, err := fs.Stat("/testDir1")
	if err != nil {
		t.Error(err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
	if !info.Mode().IsDir() {
		t.Error("FileMode is not directory")
	}
	info, err = fs.Stat("/sub/testDir2")
	if err != nil {
		t.Error(err)
	}
	if !info.IsDir() {
		t.Error("should be a directory")
	}
	if !info.Mode().IsDir() {
		t.Error("FileMode is not directory")
	}
}

func TestMemFsUnexpectedEOF(t *testing.T) {
	t.Parallel()

	fs := New()

	if err := fsutil.WriteFile(fs, "file.txt", []byte("abc"), 0777); err != nil {
		t.Fatal(err)
	}

	f, err := fs.Open("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Seek beyond the end.
	fseek, ok := f.(io.Seeker)
	if !ok {
		t.Fatal("file not a io.Seeker")
	}
	_, err = fseek.Seek(512, 0)
	if err != nil {
		t.Fatal(err)
	}

	buff := make([]byte, 256)
	_, err = io.ReadAtLeast(f, buff, 256)

	if err != io.ErrUnexpectedEOF {
		t.Fatal("Expected ErrUnexpectedEOF")
	}
}

func TestMemFsChmod(t *testing.T) {
	t.Parallel()

	fs := New()
	const file = "hello"
	if err := fs.Mkdir(file, 0700); err != nil {
		t.Fatal(err)
	}

	info, err := fs.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().String() != "drwx------" {
		t.Fatal("mkdir failed to create a directory: mode =", info.Mode())
	}

	err = fs.Chmod(file, 0)
	if err != nil {
		t.Error("Failed to run chmod:", err)
	}

	info, err = fs.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().String() != "d---------" {
		t.Error("chmod should not change file type. New mode =", info.Mode())
	}
}

// can't use Mkdir to get around which permissions we're allowed to set
func TestMemFsMkdirModeIllegal(t *testing.T) {
	t.Parallel()

	fs := New()
	err := fs.Mkdir("/a", os.ModeSocket|0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := fs.Stat("/a")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(os.ModeDir|0755) {
		t.Fatalf("should not be able to use Mkdir to set illegal mode: %s", info.Mode().String())
	}
}

// can't use OpenFile to get around which permissions we're allowed to set
func TestMemFsOpenFileModeIllegal(t *testing.T) {
	t.Parallel()

	fs := New()
	file, err := fs.OpenFile("/a", os.O_CREATE, os.ModeSymlink|0644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := fs.Stat("/a")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != os.FileMode(0644) {
		t.Fatalf("should not be able to use OpenFile to set illegal mode: %s", info.Mode().String())
	}
}
