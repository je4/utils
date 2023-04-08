package zipasfolder

import (
	"io/fs"
	"time"
)

func NewZIPFSFileInfoDir(name string) *ZIPFSFileInfoDir {
	return &ZIPFSFileInfoDir{
		name: name,
	}
}

type ZIPFSFileInfoDir struct {
	name string
}

func (zfsfid *ZIPFSFileInfoDir) Name() string {
	return zfsfid.name
}

func (zfsfid *ZIPFSFileInfoDir) Size() int64 {
	return 0
}

func (zfsfid *ZIPFSFileInfoDir) Mode() fs.FileMode {
	return fs.ModeDir | 0755
}

func (zfsfid *ZIPFSFileInfoDir) ModTime() time.Time {
	return time.Time{}
}

func (zfsfid *ZIPFSFileInfoDir) IsDir() bool {
	return true
}

func (zfsfid *ZIPFSFileInfoDir) Sys() any {
	return nil
}

func NewZIPFSFileInfoFile(name string, size int64) *ZIPFSFileInfoFile {
	return &ZIPFSFileInfoFile{
		name: name,
		size: size,
	}
}

type ZIPFSFileInfoFile struct {
	name string
	size int64
}

func (zfsfid *ZIPFSFileInfoFile) Name() string {
	return zfsfid.name
}

func (zfsfid *ZIPFSFileInfoFile) Size() int64 {
	return zfsfid.size
}

func (zfsfid *ZIPFSFileInfoFile) Mode() fs.FileMode {
	return 0644
}

func (zfsfid *ZIPFSFileInfoFile) ModTime() time.Time {
	return time.Time{}
}

func (zfsfid *ZIPFSFileInfoFile) IsDir() bool {
	return false
}

func (zfsfid *ZIPFSFileInfoFile) Sys() any {
	return nil
}

var (
	_ fs.FileInfo = &ZIPFSFileInfoDir{}
)
