// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memfs

import (
	"io/fs"
	"sort"
)

type Dir interface {
	Len() int
	Names() []string
	Files() []*FileData
	Add(*FileData)
	Remove(*FileData)
}

type dirEntry struct {
	name  string
	isDir bool
	typ   fs.FileMode
	info  fs.FileInfo
}

func (e dirEntry) Name() string {
	return e.name
}

func (e dirEntry) IsDir() bool {
	return e.isDir
}

func (e dirEntry) Type() fs.FileMode {
	return e.typ
}

func (e dirEntry) Info() (fs.FileInfo, error) {
	return e.info, nil
}

func RemoveFromMemDir(dir *FileData, f *FileData) {
	dir.memDir.Remove(f)
}

func AddToMemDir(dir *FileData, f *FileData) {
	dir.memDir.Add(f)
}

func InitializeDir(d *FileData) {
	if d.memDir == nil {
		d.dir = true
		d.memDir = &DirMap{}
	}
}

type DirMap map[string]*FileData

func (m DirMap) Len() int           { return len(m) }
func (m DirMap) Add(f *FileData)    { m[f.name] = f }
func (m DirMap) Remove(f *FileData) { delete(m, f.name) }
func (m DirMap) Files() (files []*FileData) {
	for _, f := range m {
		files = append(files, f)
	}
	sort.Sort(filesSorter(files))
	return files
}

// implement sort.Interface for []*FileData
type filesSorter []*FileData

func (s filesSorter) Len() int           { return len(s) }
func (s filesSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s filesSorter) Less(i, j int) bool { return s[i].name < s[j].name }

func (m DirMap) Names() (names []string) {
	for x := range m {
		names = append(names, x)
	}
	return names
}
