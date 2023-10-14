package mountfs

import (
	"io/fs"

	"tractor.dev/toolkit-go/engine/fs/readonlyfs"
	"tractor.dev/toolkit-go/engine/fs/unionfs"
	"tractor.dev/toolkit-go/engine/fs/workingpathfs"
)

type Modifier func(fs.FS, string, fs.FS) fs.FS

func Union() Modifier {
	return func(fsys fs.FS, root string, mount fs.FS) fs.FS {
		if root != "" && root != "." {
			fsys = workingpathfs.New(fsys, root)
		}
		return unionfs.New(fsys, mount)
	}
}

// func CopyOnWrite(overlay fs.FS) Modifier {
// 	return func(fsys fs.FS, root string, mount fs.FS) fs.FS {
// 		return afero.NewCopyOnWriteFs(mount, overlay)
// 	}
// }

// func CacheOnRead(cache fs.FS, cacheTime time.Duration) Modifier {
// 	return func(fsys fs.FS, root string, mount fs.FS) fs.FS {
// 		return afero.NewCacheOnReadFs(mount, cache, cacheTime)
// 	}
// }

func ReadOnly() Modifier {
	return func(fsys fs.FS, root string, mount fs.FS) fs.FS {
		return readonlyfs.New(mount)
	}
}
