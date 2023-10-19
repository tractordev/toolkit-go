package fstest

import (
	"io/fs"
	"path/filepath"
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fsutil"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func ReadFS(t *testing.T, fsys fs.FS) map[string]string {
	fsmap := map[string]string{}
	must(t, fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			fsmap[path+"/"] = ""
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		fsmap[path] = string(b)
		return nil
	}))
	return fsmap
}

func CheckFS(t *testing.T, fsys fs.FS, expect map[string]string) {
	fsmap := ReadFS(t, fsys)
	for path, contents := range expect {
		c, ok := fsmap[path]
		if !ok {
			t.Fatal("path not found:", path)
		}
		if c != contents {
			t.Fatalf("unexpected contents for %s: got %#v, want %#v", path, c, contents)
		}
	}
}

func WriteFS(t *testing.T, fsys fs.FS, fsmap map[string]string) {
	for path, contents := range fsmap {
		must(t, fsutil.MkdirAll(fsys, filepath.Dir(path), 0755))
		must(t, fsutil.WriteFile(fsys, path, []byte(contents), 0644))
	}
}
