package unionfs

import (
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fstest"
	"tractor.dev/toolkit-go/engine/fs/memfs"
)

func TestUnionFS(t *testing.T) {

	base := memfs.New()
	overlay := memfs.New()

	fstest.WriteFS(t, base, map[string]string{
		"all":      "base",
		"sub/base": "base",
	})

	fstest.WriteFS(t, overlay, map[string]string{
		"all":         "overlay",
		"sub/overlay": "overlay",
	})

	fsys := New(base, overlay)

	fstest.CheckFS(t, fsys, map[string]string{
		"all":         "overlay",
		"sub/base":    "base",
		"sub/overlay": "overlay",
	})

}
