package concurrentWriter

import (
	"github.com/pkg/errors"
	"io"
)

func NewGenericCopyRunner(writer io.Writer, name string) *GenericCopyRunner {
	return &GenericCopyRunner{
		writer: writer,
		name:   name,
	}
}

type GenericCopyRunner struct {
	writer io.Writer
	name   string
	err    error
}

func (w *GenericCopyRunner) Do(reader io.Reader, done chan bool) {
	// we should end in all cases
	defer func() {
		done <- true
	}()

	if _, err := io.Copy(w.writer, reader); err != nil {
		w.err = errors.Wrapf(err, "%s cannot copy", w.name)
	}
}

func (w *GenericCopyRunner) GetError() error {
	return w.err
}

func (w *GenericCopyRunner) GetName() string {
	return w.name
}
