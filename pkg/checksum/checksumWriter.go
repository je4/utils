package checksum

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/concurrentWriter"
	"io"
)

type ChecksumWriter struct {
	writer *concurrentWriter.ConcurrentWriter
}

func NewChecksumWriter(checksums []DigestAlgorithm, writers ...io.Writer) (*ChecksumWriter, error) {
	var runners = []concurrentWriter.WriterRunner{}
	for _, alg := range checksums {
		runner, err := NewWriterRunnerChecksum(alg)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create runner for '%s'", alg)
		}
		runners = append(runners, runner)
	}
	writer := concurrentWriter.NewConcurrentWriter(runners, writers...)
	c := &ChecksumWriter{
		writer: writer,
	}
	return c, nil
}

func (c *ChecksumWriter) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c *ChecksumWriter) Close() error {
	if err := c.writer.Close(); err != nil {
		return errors.Wrap(err, "cannot close concurrent writer")
	}
	return nil
}

func (c *ChecksumWriter) GetChecksums() (map[DigestAlgorithm]string, error) {
	var result = map[DigestAlgorithm]string{}
	for _, runner := range c.writer.GetRunners() {
		r := runner.(*WriterRunnerChecksum)
		digest, err := r.GetDigest()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get digest from '%s'", r.GetAlgorithm())
		}
		result[r.GetAlgorithm()] = digest
	}
	return result, nil
}

var (
	_ io.WriteCloser = (*ChecksumWriter)(nil)
)
