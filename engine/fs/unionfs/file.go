package unionfs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
)

// The UnionFile implements the afero.File interface and will be returned
// when reading a directory present at least in the overlay or opening a file
// for writing.
//
// The calls to
// Readdir() and Readdirnames() merge the file os.FileInfo / names from the
// base and the overlay - for files present in both layers, only those
// from the overlay will be used.
//
// When opening files for writing (Create() / OpenFile() with the right flags)
// the operations will be done in both layers, starting with the overlay. A
// successful read in the overlay will move the cursor position in the base layer
// by the number of bytes read.
type File struct {
	Base   fs.File
	Layer  fs.File
	Merger DirsMerger
	off    int
	files  []fs.DirEntry
}

func (f *File) Close() error {
	// first close base, so we have a newer timestamp in the overlay. If we'd close
	// the overlay first, we'd get a cacheStale the next time we access this file
	// -> cache would be useless ;-)
	if f.Base != nil {
		f.Base.Close()
	}
	if f.Layer != nil {
		return f.Layer.Close()
	}
	return fs.ErrInvalid
}

func (f *File) Read(s []byte) (int, error) {
	if f.Layer != nil {
		n, err := f.Layer.Read(s)
		if (err == nil || err == io.EOF) && f.Base != nil {
			// advance the file position also in the base file, the next
			// call may be a write at this position (or a seek with SEEK_CUR)
			fseek, ok := f.Base.(io.Seeker)
			if !ok {
				return n, fmt.Errorf("unable to seek file")
			}
			if _, seekErr := fseek.Seek(int64(n), os.SEEK_CUR); seekErr != nil {
				// only overwrite err in case the seek fails: we need to
				// report an eventual io.EOF to the caller
				err = seekErr
			}
		}
		return n, err
	}
	if f.Base != nil {
		return f.Base.Read(s)
	}
	return 0, fs.ErrInvalid
}

func (f *File) ReadAt(s []byte, o int64) (int, error) {
	if f.Layer != nil {
		freadat, ok := f.Layer.(io.ReaderAt)
		if !ok {
			return 0, fmt.Errorf("unable to readat file")
		}
		n, err := freadat.ReadAt(s, o)
		if (err == nil || err == io.EOF) && f.Base != nil {
			fseek, ok := f.Base.(io.Seeker)
			if !ok {
				return n, fmt.Errorf("unable to seek file")
			}
			_, err = fseek.Seek(o+int64(n), os.SEEK_SET)
		}
		return n, err
	}
	if f.Base != nil {
		freadat, ok := f.Layer.(io.ReaderAt)
		if !ok {
			return 0, fmt.Errorf("unable to readat file")
		}
		return freadat.ReadAt(s, o)
	}
	return 0, fs.ErrInvalid
}

func (f *File) Seek(o int64, w int) (pos int64, err error) {
	if f.Layer != nil {
		fseek, ok := f.Layer.(io.Seeker)
		if !ok {
			return 0, fmt.Errorf("unable to seek file")
		}
		pos, err = fseek.Seek(o, w)
		if (err == nil || err == io.EOF) && f.Base != nil {
			fseek, ok := f.Base.(io.Seeker)
			if !ok {
				return 0, fmt.Errorf("unable to seek file")
			}
			_, err = fseek.Seek(o, w)
		}
		return pos, err
	}
	if f.Base != nil {
		fseek, ok := f.Base.(io.Seeker)
		if !ok {
			return 0, fmt.Errorf("unable to seek file")
		}
		return fseek.Seek(o, w)
	}
	return 0, fs.ErrInvalid
}

func (f *File) Write(s []byte) (n int, err error) {
	if f.Layer != nil {
		fw, ok := f.Layer.(io.Writer)
		if !ok {
			return 0, fs.ErrPermission
		}
		n, err = fw.Write(s)
		if err == nil && f.Base != nil { // hmm, do we have fixed size files where a write may hit the EOF mark?
			fw, ok := f.Base.(io.Writer)
			if !ok {
				return 0, fs.ErrPermission
			}
			_, err = fw.Write(s)
		}
		return n, err
	}
	if f.Base != nil {
		fw, ok := f.Base.(io.Writer)
		if !ok {
			return 0, fs.ErrPermission
		}
		return fw.Write(s)
	}
	return 0, fs.ErrInvalid
}

func (f *File) WriteAt(s []byte, o int64) (n int, err error) {
	if f.Layer != nil {
		fwriteat, ok := f.Layer.(io.WriterAt)
		if !ok {
			return 0, fmt.Errorf("unable to writeat file")
		}
		n, err = fwriteat.WriteAt(s, o)
		if err == nil && f.Base != nil {
			fwriteat, ok := f.Base.(io.WriterAt)
			if !ok {
				return 0, fmt.Errorf("unable to writeat file")
			}
			_, err = fwriteat.WriteAt(s, o)
		}
		return n, err
	}
	if f.Base != nil {
		fwriteat, ok := f.Base.(io.WriterAt)
		if !ok {
			return 0, fmt.Errorf("unable to writeat file")
		}
		return fwriteat.WriteAt(s, o)
	}
	return 0, fs.ErrInvalid
}

// DirsMerger is how UnionFile weaves two directories together.
// It takes the FileInfo slices from the layer and the base and returns a
// single view.
type DirsMerger func(lofi, bofi []fs.DirEntry) ([]fs.DirEntry, error)

var defaultUnionMergeDirsFn = func(lofi, bofi []fs.DirEntry) ([]fs.DirEntry, error) {
	var files = make(map[string]fs.DirEntry)

	for _, fi := range lofi {
		files[fi.Name()] = fi
	}

	for _, fi := range bofi {
		if _, exists := files[fi.Name()]; !exists {
			files[fi.Name()] = fi
		}
	}

	rfi := make([]fs.DirEntry, len(files))

	i := 0
	for _, fi := range files {
		rfi[i] = fi
		i++
	}

	return rfi, nil

}

// Readdir will weave the two directories together and
// return a single view of the overlayed directories.
// At the end of the directory view, the error is io.EOF if c > 0.
func (f *File) ReadDir(c int) (ofi []fs.DirEntry, err error) {
	var merge DirsMerger = f.Merger
	if merge == nil {
		merge = defaultUnionMergeDirsFn
	}

	if f.off == 0 {
		var lfi []fs.DirEntry
		if f.Layer != nil {
			df, ok := f.Layer.(fs.ReadDirFile)
			if !ok {
				return nil, fmt.Errorf("unable to readdir file")
			}
			lfi, err = df.ReadDir(-1)
			if err != nil {
				return nil, err
			}
		}

		var bfi []fs.DirEntry
		if f.Base != nil {
			df, ok := f.Base.(fs.ReadDirFile)
			if !ok {
				return nil, fmt.Errorf("unable to readdir file")
			}
			bfi, err = df.ReadDir(-1)
			if err != nil {
				return nil, err
			}

		}
		merged, err := merge(lfi, bfi)
		if err != nil {
			return nil, err
		}
		f.files = append(f.files, merged...)
	}
	files := f.files[f.off:]

	if c <= 0 {
		return files, nil
	}

	if len(files) == 0 {
		return nil, io.EOF
	}

	if c > len(files) {
		c = len(files)
	}

	defer func() { f.off += c }()
	return files[:c], nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	if f.Layer != nil {
		return f.Layer.Stat()
	}
	if f.Base != nil {
		return f.Base.Stat()
	}
	return nil, fs.ErrInvalid
}

func (f *File) Sync() (err error) {
	if f.Layer != nil {
		fsync, ok := f.Layer.(interface{ Sync() error })
		if !ok {
			return fmt.Errorf("unable to sync file")
		}
		err = fsync.Sync()
		if err == nil && f.Base != nil {
			fsync, ok := f.Base.(interface{ Sync() error })
			if !ok {
				return fmt.Errorf("unable to sync file")
			}
			err = fsync.Sync()
		}
		return err
	}
	if f.Base != nil {
		fsync, ok := f.Base.(interface{ Sync() error })
		if !ok {
			return fmt.Errorf("unable to sync file")
		}
		return fsync.Sync()
	}
	return fs.ErrInvalid
}

func (f *File) Truncate(s int64) (err error) {
	if f.Layer != nil {
		ft, ok := f.Layer.(interface{ Truncate(int64) error })
		if !ok {
			return fmt.Errorf("unable to truncate file")
		}
		err = ft.Truncate(s)
		if err == nil && f.Base != nil {
			ft, ok := f.Base.(interface{ Truncate(int64) error })
			if !ok {
				return fmt.Errorf("unable to truncate file")
			}
			err = ft.Truncate(s)
		}
		return err
	}
	if f.Base != nil {
		ft, ok := f.Base.(interface{ Truncate(int64) error })
		if !ok {
			return fmt.Errorf("unable to truncate file")
		}
		return ft.Truncate(s)
	}
	return fs.ErrInvalid
}

// func copyToLayer(base fs.FS, layer fs.FS, name string) error {
// 	bfh, err := base.Open(name)
// 	if err != nil {
// 		return err
// 	}
// 	defer bfh.Close()

// 	// First make sure the directory exists
// 	exists, err := vfs.Exists(layer, filepath.Dir(name))
// 	if err != nil {
// 		return err
// 	}
// 	if !exists {
// 		err = layer.MkdirAll(filepath.Dir(name), 0777) // FIXME?
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// Create the file on the overlay
// 	lfh, err := layer.Create(name)
// 	if err != nil {
// 		return err
// 	}
// 	n, err := io.Copy(lfh, bfh)
// 	if err != nil {
// 		// If anything fails, clean up the file
// 		layer.Remove(name)
// 		lfh.Close()
// 		return err
// 	}

// 	bfi, err := bfh.Stat()
// 	if err != nil || bfi.Size() != n {
// 		layer.Remove(name)
// 		lfh.Close()
// 		return syscall.EIO
// 	}

// 	err = lfh.Close()
// 	if err != nil {
// 		layer.Remove(name)
// 		lfh.Close()
// 		return err
// 	}
// 	return layer.Chtimes(name, bfi.ModTime(), bfi.ModTime())
// }
