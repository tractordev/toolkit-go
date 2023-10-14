package fs

import (
	iofs "io/fs"
)

var (
	ErrInvalid    = iofs.ErrInvalid
	ErrPermission = iofs.ErrPermission
	ErrExist      = iofs.ErrExist
	ErrNotExist   = iofs.ErrNotExist
	ErrClosed     = iofs.ErrClosed

	SkipAll = iofs.SkipAll
	SkipDir = iofs.SkipDir
)

var (
	FormatDirEntry     = iofs.FormatDirEntry
	FormatFileInfo     = iofs.FormatFileInfo
	Glob               = iofs.Glob
	ReadFile           = iofs.ReadFile
	ValidPath          = iofs.ValidPath
	WalkDir            = iofs.WalkDir
	FileInfoToDirEntry = iofs.FileInfoToDirEntry
	ReadDir            = iofs.ReadDir
	Sub                = iofs.Sub
	Stat               = iofs.Stat
)

type (
	DirEntry    = iofs.DirEntry
	FS          = iofs.FS
	File        = iofs.File
	FileInfo    = iofs.FileInfo
	FileMode    = iofs.FileMode
	GlobFS      = iofs.GlobFS
	PathError   = iofs.PathError
	ReadDirFS   = iofs.ReadDirFS
	ReadDirFile = iofs.ReadDirFile
	ReadFileFS  = iofs.ReadFileFS
	StatFS      = iofs.StatFS
	SubFS       = iofs.SubFS
	WalkDirFunc = iofs.WalkDirFunc
)
