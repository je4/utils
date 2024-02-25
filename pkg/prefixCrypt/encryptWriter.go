package prefixCrypt

import (
	"emperror.dev/errors"
	"io"
)

func NewEncWriter(w io.Writer, encrypt Encrypter) *EncWriter {
	return &EncWriter{
		w:       w,
		buf:     []byte{},
		encrypt: encrypt,
	}
}

type EncWriter struct {
	w       io.Writer
	buf     []byte
	encrypt Encrypter
	offset  int64
}

func (e *EncWriter) Close() error {
	if len(e.buf) > 0 {
		enc, err := e.encrypt.Encrypt(e.buf)
		if err != nil {
			return errors.Wrap(err, "cannot encrypt buffer")
		}
		if _, err := e.w.Write(enc); err != nil {
			return errors.WithStack(err)
		}
		e.buf = nil
	}
	return nil
}

func (e *EncWriter) Write(p []byte) (n int, err error) {
	// rest size of the buffer
	bufferCap := max(SIZE-e.offset, 0)
	// number of bytes to write to buffer
	bufferWrite := min(int64(len(p)), bufferCap)
	if bufferWrite > 0 {
		if e.buf == nil {
			e.buf = make([]byte, 0, SIZE)
		}
		e.buf = append(e.buf, p[:bufferWrite]...)
		n += int(bufferWrite)
		p = p[bufferWrite:]
		e.offset += bufferWrite
	}
	// if buffer is full, encrypt and write it before any other data is written
	if len(p) > 0 && e.buf != nil {
		enc, err := e.encrypt.Encrypt(e.buf)
		e.buf = nil
		if err != nil {
			return 0, errors.Wrap(err, "cannot encrypt buffer")
		}
		n, err = e.w.Write(enc)
		if err != nil {
			return n, errors.WithStack(err)
		}
	}
	x, err := e.w.Write(p)
	if err != nil {
		return n, errors.WithStack(err)
	}
	n += x
	e.offset += int64(x)
	return
}

var _ io.WriteCloser = (*EncWriter)(nil)
