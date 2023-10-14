package watchfs

import (
	"fmt"
	"io/fs"
)

type WatchFS interface {
	fs.FS
	Watch(name string, cfg *Config) (*Watch, error)
}

func WatchFile(fsys fs.FS, name string, cfg *Config) (*Watch, error) {
	if fsys, ok := fsys.(WatchFS); ok {
		return fsys.Watch(name, cfg)
	}

	return nil, fmt.Errorf("watch %s: operation not supported", name)
}
