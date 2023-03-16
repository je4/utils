package indexer

import (
	"emperror.dev/errors"
	"io"
	"net/http"
)

type MimeReader struct {
	io.Reader
	buffer      []byte
	contentType string
}

func NewMimeReader(r io.Reader) (*MimeReader, error) {
	mr := &MimeReader{
		Reader: r,
		buffer: make([]byte, 512),
	}
	return mr, mr.Init()
}

func (mr *MimeReader) Init() error {
	n, err := mr.Reader.Read(mr.buffer)
	if err != nil {
		return errors.Wrap(err, "failed to read head")
	}
	mr.buffer = mr.buffer[:n]
	mr.contentType = http.DetectContentType(mr.buffer)
	return nil
}

func (mr *MimeReader) DetectContentType() (string, error) {
	return mr.contentType, nil
}

func (mr *MimeReader) Read(p []byte) (n int, err error) {
	if len(mr.buffer) > 0 {
		n = copy(p, mr.buffer)
		mr.buffer = mr.buffer[n:]
		return
	}
	return mr.Reader.Read(p)
}
