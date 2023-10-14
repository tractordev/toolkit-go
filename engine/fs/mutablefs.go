package fs

import "time"

type MutableFS interface {
	StatFS

	Chmod(name string, mode FileMode) error
	Chown(name string, uid, gid int) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Create(name string) (File error)
	Mkdir(name string, perm FileMode) error
	MkdirAll(path string, perm FileMode) error
	OpenFile(name string, flag int, perm FileMode) (File, error)
	Remove(name string) error
	RemoveAll(path string) error
	Rename(oldname, newname string) error
}
