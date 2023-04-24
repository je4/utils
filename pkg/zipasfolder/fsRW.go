package zipasfolder

import (
	"io"
	"io/fs"
)

type FileW interface {
	io.WriteCloser
}

type FSRW interface {
	fs.StatFS
	Create(path string) (FileW, error)
	MkDir(path string) error
}

type FSRWClose interface {
	Close() error
}
