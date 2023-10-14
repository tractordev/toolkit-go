package makefs

import (
	"bytes"
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fstest"
	"tractor.dev/toolkit-go/engine/fs/memfs"
)

type Apply struct {
	Fn func([]byte) []byte
	b  bytes.Buffer
}

func (x *Apply) Reset() {}

func (x *Apply) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := copy(dst, x.Fn(src))
	return n, n, nil
}

func TestMountOpener(t *testing.T) {
	base := memfs.New()

	fstest.WriteFS(t, base, map[string]string{
		"src/hello": "Hello World",
	})

	fsys := MountOpener(base, "dst/upper", TransformFrom(base, "src/hello", &Apply{Fn: bytes.ToUpper}))
	fsys = MountOpener(fsys, "dst/lower", TransformFrom(fsys, "src/hello", &Apply{Fn: bytes.ToLower}))

	fstest.CheckFS(t, fsys, map[string]string{
		"src/hello": "Hello World",
		"dst/upper": "HELLO WORLD",
		"dst/lower": "hello world",
	})

	fsys = MountOpener(fsys, "dst/bundle", MakeOpener(fsys, func(files []OpenFile) ([]byte, error) {
		var buf bytes.Buffer
		for _, f := range files {
			buf.ReadFrom(f.File)
		}
		return buf.Bytes(), nil
	}, "dst/*"))

	fsys = MountOpener(base, "gen", MakeOpener(fsys, func(files []OpenFile) ([]byte, error) {
		return []byte("pretend generated"), nil
	}))

	fstest.CheckFS(t, fsys, map[string]string{
		"src/hello":  "Hello World",
		"dst/upper":  "HELLO WORLD",
		"dst/lower":  "hello world",
		"dst/bundle": "hello worldHELLO WORLD",
		"gen":        "pretend generated",
	})

}
