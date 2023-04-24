package zipasfolder

import (
	"io/fs"
	"os"
	"path/filepath"
)

func NewDummyOSRW(dir string) FSRW {
	return &dummyOSRW{
		dir: dir,
	}
}

type dummyOSRW struct {
	dir string
}

func (d *dummyOSRW) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(d.dir, name))
}

func (d *dummyOSRW) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(d.dir, name))
}

func (d *dummyOSRW) Create(path string) (FileW, error) {
	return os.Create(filepath.Join(d.dir, path))
}

func (d *dummyOSRW) MkDir(path string) error {
	return os.Mkdir(filepath.Join(d.dir, path), 0777)
}

func (d *dummyOSRW) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(d.dir, name))
}

func (d *dummyOSRW) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(d.dir, name))
}
