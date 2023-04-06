package zipasfolder

import (
	"archive/zip"
	"github.com/pkg/errors"
	"io/fs"
	"sync"
)

func NewZIPFS(zipReader *zip.Reader, zipFile fs.File) *ZIPFS {
	return &ZIPFS{
		zipReader: zipReader,
		zipFile:   zipFile,
		lock:      &sync.Mutex{},
	}
}

type ZIPFS struct {
	zipReader *zip.Reader
	zipFile   fs.File
	lock      *sync.Mutex
}

func (zipFS *ZIPFS) Close() error {
	return zipFS.zipFile.Close()
}

func (zipFS *ZIPFS) Open(name string) (fs.File, error) {
	zipFS.lock.Lock()
	//	defer zipFS.lock.Unlock()
	for _, f := range zipFS.zipReader.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				zipFS.lock.Unlock()
				return nil, errors.WithStack(err)
			}
			return NewFile(f.FileInfo(), rc, zipFS.lock), nil
		}
	}
	return nil, fs.ErrNotExist
}
