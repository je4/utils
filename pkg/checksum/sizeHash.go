package checksum

import (
	"hash"
	"strconv"
)

type sizeHash struct {
	size uint64
}

func NewSizeHash() hash.Hash {
	return &sizeHash{}
}

func (s *sizeHash) Write(p []byte) (n int, err error) {
	n = len(p)
	s.size += uint64(n)
	return n, nil
}

func (s *sizeHash) Sum(b []byte) []byte {
	sizeStr := strconv.FormatInt(int64(s.size), 10)
	return append(b, []byte(sizeStr)...)
}

func (s *sizeHash) Reset() {
	s.size = 0
}

func (s *sizeHash) Size() int {
	return 8
}

func (s *sizeHash) BlockSize() int {
	return 1
}
