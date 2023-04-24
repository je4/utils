package zipasfolder

import (
	"io/fs"
	"path/filepath"
)

type subFS struct {
	fsys FSRW
	dir  string
}

func NewSubFS(fsys FSRW, dir string) *subFS {
	return &subFS{
		fsys: fsys,
		dir:  dir,
	}
}

func (sfs *subFS) Open(name string) (fs.File, error) {
	return sfs.fsys.Open(filepath.ToSlash(filepath.Join(sfs.dir, name)))
}

func (sfs *subFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(sfs.fsys, filepath.ToSlash(filepath.Join(sfs.dir, name)))
}

func (sfs *subFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(sfs.fsys, filepath.ToSlash(filepath.Join(sfs.dir, name)))
}

func (sfs *subFS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(sfs.fsys, filepath.ToSlash(filepath.Join(sfs.dir, name)))
}

func (sfs *subFS) Sub(dir string) (FSRW, error) {
	return NewSubFS(sfs.fsys, filepath.ToSlash(filepath.Join(sfs.dir, dir))), nil
}

func (sfs *subFS) Create(path string) (FileW, error) {
	return sfs.fsys.Create(filepath.ToSlash(filepath.Join(sfs.dir, path)))
}

func (sfs *subFS) MkDir(path string) error {
	return sfs.fsys.MkDir(filepath.ToSlash(filepath.Join(sfs.dir, path)))
}
