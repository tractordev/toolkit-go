package fs

import (
	"embed"
	"os"
	"path/filepath"
	"runtime"

	"tractor.dev/toolkit-go/engine/fs/watchfs"
)

// LiveDir uses the filename of the calling source file and sees if it exists on the system
// to return a watchable version of that directory, otherwise it returns the assets passed in.
func LiveDir(assets embed.FS) FS {
	_, filename, _, _ := runtime.Caller(1)
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return assets
	}
	return watchfs.New(os.DirFS(filepath.Dir(filename)))
}
