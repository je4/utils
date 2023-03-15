package indexer

import (
	"emperror.dev/errors"
	"io"
)

type WriteIgnoreCloser struct {
	io.WriteCloser
}

func NewWriteIgnoreCloser(wc io.WriteCloser) *WriteIgnoreCloser {
	return &WriteIgnoreCloser{wc}
}

func (wcl *WriteIgnoreCloser) Close() error {
	return nil
}

func (wcl *WriteIgnoreCloser) ForceClose() error {
	return errors.WithStack(wcl.WriteCloser.Close())
}

type ReadIgnoreCloser struct {
	io.ReadCloser
}

func NewReadIgnoreCloser(rc io.ReadCloser) *ReadIgnoreCloser {
	return &ReadIgnoreCloser{rc}
}

func (wcl *ReadIgnoreCloser) Close() error {
	return nil
}

func (wcl *ReadIgnoreCloser) ForceClose() error {
	return errors.WithStack(wcl.ReadCloser.Close())
}
