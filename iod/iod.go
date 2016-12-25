package iod

import "io"

type RW interface {
	io.ReaderAt
	io.WriterAt
	io.Closer
}

type ORW struct {
	rw  RW
	off int64
}

func NewORW(rw RW, off int64) *ORW {
	return &ORW{rw, off}
}

func (s *ORW) ReadAt(p []byte, off int64) (n int, err error) {
	return s.rw.ReadAt(p, s.off+off)
}

func (s *ORW) WriteAt(p []byte, off int64) (n int, err error) {
	return s.rw.WriteAt(p, s.off+off)
}

func (s *ORW) Close() error {
	return s.rw.Close()
}
