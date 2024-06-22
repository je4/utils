package indexer

import (
	"bytes"
	"emperror.dev/errors"
	"io"
	"net/http"
)

const bufSize = 512

type MimeReader struct {
	io.Reader
	buffer      []byte
	contentType string
}

func NewMimeReader(r io.Reader) (*MimeReader, error) {
	reader := r // bufio.NewReaderSize(r, bufSize)
	mr := &MimeReader{
		Reader: reader,
		buffer: make([]byte, 0, bufSize),
	}
	return mr, mr.Init()
}

func (mr *MimeReader) Init() error {
	var n int64
	var err error
	n, err = io.CopyN(bytes.NewBuffer(mr.buffer), mr.Reader, int64(bufSize))
	//	n, err = mr.Reader.Read(mr.buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			mr.contentType = "application/octet-stream"
			mr.buffer = make([]byte, 0, bufSize)
			return nil
		}
		return errors.Wrap(err, "failed to read head")
	}
	mr.buffer = mr.buffer[:n]
	mr.contentType = http.DetectContentType(mr.buffer)
	if n == 0 {
		mr.contentType = "application/octet-stream"
	}
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
