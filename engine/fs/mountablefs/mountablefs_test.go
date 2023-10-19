package mountablefs

import (
	"io/fs"
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fstest"
	"tractor.dev/toolkit-go/engine/fs/memfs"
)

func TestMountableFS(t *testing.T) {
    host := memfs.New()
    mnt := memfs.New()

    fstest.WriteFS(t, host, map[string]string{
        "all":             "host",
        "mount/host-data": "host",
    })

    fstest.WriteFS(t, mnt, map[string]string{
        "all2":         "mounted",
        "rickroll.mpv": "mounted",
    })

    fsys := New(host)
    if err := fsys.Mount(mnt, "mount"); err != nil {
        t.Fatal(err)
    }

    fstest.CheckFS(t, fsys, map[string]string{
        "all": "host",
        "mount/all2": "mounted",
        "mount/rickroll.mpv": "mounted",
    })

    if _, err := fsys.Open("mount/host-data"); err == nil {
        t.Fatalf("expected file %s to be masked by mount: got nil, expected %v", "mount/host-data", fs.ErrNotExist)
    }

    if err := fsys.Unmount("all"); err == nil {
        t.Fatal("expected error when attempting to Unmount a non-mountpoint, got nil")
    }

    if err := fsys.Unmount("mount"); err != nil {
        t.Fatal(err)
    }

    fstest.CheckFS(t, fsys, map[string]string{
        "all":             "host",
        "mount/host-data": "host",
    })

    if _, err := fsys.Open("mount/rickroll.mpv"); err == nil {
        t.Fatalf("unexpected file %s: expected error %v", "mount/rickroll.mpv", fs.ErrNotExist)
    }
}
