package watchfs

import (
	"errors"
	"fmt"
	"io/fs"
)

type WatchFS interface {
	fs.FS
	// Wrapper Filesystems should check their children for implementations before providing their own.
	// Return `errors.ErrUnsupported` to signal that parent Filesystems should use their implementation instead.
	Watch(name string, cfg *Config) (*Watch, error)
}

func WatchFile(fsys fs.FS, name string, cfg *Config) (*Watch, error) {
	if fsys, ok := fsys.(WatchFS); ok {
		return fsys.Watch(name, cfg)
	}

	return nil, fmt.Errorf("watch %s: %w", name, errors.ErrUnsupported)
}
