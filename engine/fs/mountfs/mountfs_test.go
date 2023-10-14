package mountfs

import (
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fstest"
	"tractor.dev/toolkit-go/engine/fs/memfs"
)

func TestMountFS(t *testing.T) {
	base := memfs.New()
	mount := memfs.New()

	fstest.WriteFS(t, base, map[string]string{
		"basefile":       "basefile",
		"mount/file":     "basefile",
		"mount/basefile": "basefile",
	})

	fstest.WriteFS(t, mount, map[string]string{
		"sub/file": "mountfile",
		"file":     "mountfile",
	})

	fsys := New(base, "mount", mount)

	fstest.CheckFS(t, fsys, map[string]string{
		"basefile":       "basefile",
		"mount/file":     "mountfile",
		"mount/sub/file": "mountfile",
	})

}

func TestMountFSUnionRoot(t *testing.T) {
	base := memfs.New()
	mnt := memfs.New()

	fstest.WriteFS(t, base, map[string]string{
		"basefile": "basefile",
		"bothfile": "basefile",
	})

	fstest.WriteFS(t, mnt, map[string]string{
		"bothfile":  "mountfile",
		"mountfile": "mountfile",
	})

	fsys := New(base, "", mnt, Union())

	fstest.CheckFS(t, fsys, map[string]string{
		"basefile":  "basefile",
		"bothfile":  "mountfile",
		"mountfile": "mountfile",
	})

}

func TestMountFSUnionSub(t *testing.T) {
	base := memfs.New()
	mnt := memfs.New()

	fstest.WriteFS(t, base, map[string]string{
		"sub/basefile": "basefile",
		"sub/bothfile": "basefile",
	})

	fstest.WriteFS(t, mnt, map[string]string{
		"bothfile":  "mountfile",
		"mountfile": "mountfile",
	})

	fsys := New(base, "sub", mnt, Union())

	fstest.CheckFS(t, fsys, map[string]string{
		"sub/basefile":  "basefile",
		"sub/bothfile":  "mountfile",
		"sub/mountfile": "mountfile",
	})

}
