package indexer

import (
	"bytes"
	"io"
	"net/http"

	"emperror.dev/errors"
)

const bufSize = 512

type MimeReader struct {
	io.Reader
	buffer      []byte
	contentType string
}

// NewMimeReader creates a NewMimeReader object.
func NewMimeReader(r io.Reader) (*MimeReader, error) {
	reader := r
	mr := &MimeReader{
		Reader: reader,
	}
	return mr, mr.Init()
}

// Init copies the first `bufSize` bytes from a bufio allowing them to
// be analyzed without consuming the original bufio.
func (mr *MimeReader) Init() error {
	const defaultMime = "application/octet-stream"
	var bytesRead int64
	var err error
	buf := bytes.NewBuffer(nil)
	bytesRead, err = io.CopyN(buf, mr.Reader, int64(bufSize))
	if err != nil {
		if errors.Is(err, io.EOF) && bytesRead >= 0 {
			mr.contentType = defaultMime
			// CopyN might have been greater than the length of the stream
			// but we might still have data valuable to the caller.
			mr.Reader = io.MultiReader(bytes.NewReader(buf.Bytes()), mr.Reader)
			return nil
		}
		mr.buffer = make([]byte, 0, bufSize)
		return errors.Wrap(err, "failed to read BOF")
	}
	mr.buffer = buf.Bytes()
	mr.contentType = http.DetectContentType(mr.buffer)
	if bytesRead == 0 {
		mr.contentType = defaultMime
	}
	return nil
}

// DetectContentType returns the stored content type.
func (mr *MimeReader) DetectContentType() (string, error) {
	return mr.contentType, nil
}

func (mr *MimeReader) Read(p []byte) (n int, err error) {
	if len(mr.buffer) > 0 {
		capacity := len(p)
		n = copy(p, mr.buffer)
		mr.buffer = mr.buffer[n:]
		if n < capacity {
			h := make([]byte, capacity-n)
			n2, err := mr.Reader.Read(h)
			if err != nil {
				return n, err
			}
			n += copy(p[n:], h[:n2])
		}
		return
	}
	return mr.Reader.Read(p)
}
