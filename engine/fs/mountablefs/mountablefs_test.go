package mountablefs

import (
	"io/fs"
	"testing"

	"tractor.dev/toolkit-go/engine/fs/fstest"
	"tractor.dev/toolkit-go/engine/fs/memfs"
)

func TestMountUnmount(t *testing.T) {
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

func TestRemove(t *testing.T) {
    host := memfs.New()
    mnt := memfs.New()

    fstest.WriteFS(t, host, map[string]string{
        "A/B": "host",
        "C/D/blah": "host",
    })

    fstest.WriteFS(t, mnt, map[string]string{
        "E/F": "mounted",
        "G/H": "mounted",
    })

    fsys := New(host)
    if err := fsys.Mount(mnt, "C/D"); err != nil {
        t.Fatal(err)
    }

    fstest.CheckFS(t, fsys, map[string]string{
        "A/B": "host",
        "C/D/E/F": "mounted",
        "C/D/G/H": "mounted",
    })

    if err := fsys.Remove("A/B"); err != nil {
        t.Fatal(err)
    }

    if err := fsys.Remove("C/D/E/F"); err != nil {
        t.Fatal(err)
    }

    if err := fsys.RemoveAll("C/D/G"); err != nil {
        t.Fatal(err)
    }

    fstest.CheckFS(t, fsys, map[string]string{
        // dirs are empty
        "A/": "",
        "C/D/E/": "",
    })

    if err := fsys.Remove("A/B"); err == nil {
        t.Fatalf("expected attempt to Remove a non-existant file to fail")
    }

    if err := fsys.Remove("C/D/G"); err == nil {
        t.Fatalf("expected attempt to Remove a non-existant file to fail")
    }

    if err := fsys.Unmount("C/D"); err != nil {
        t.Fatal(err)
    }
}

func TestRename(t *testing.T) {
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

    if err := fsys.Rename("all", "none"); err != nil {
        t.Fatal(err)
    }
    
    if err := fsys.Rename("mount/all2", "mount/none2"); err != nil {
        t.Fatal(err)
    }

    if err := fsys.Rename("mount/rickroll.mpv", "rickroll.mpv"); err == nil {
        t.Fatalf("expected error when attempting to rename across filesystems")
    }

    fstest.CheckFS(t, fsys, map[string]string{
        "none": "host",
        "mount/none2": "mounted",
        "mount/rickroll.mpv": "mounted",
    })

    if err := fsys.Unmount("mount"); err != nil {
        t.Fatal(err)
    }
}

