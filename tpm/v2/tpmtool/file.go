package tpmtool

import (
	"encoding/binary"
	"io"
)

type Header struct {
	Magic       uint32
	Version     uint32
	Hierarchy   uint32
	SavedHandle uint32
	Sequence    uint64
	Length      uint16
}

type File struct {
	Header
	Blob []byte
}

func Decode(r io.Reader) (f *File, err error) {
	f = &File{}
	err = binary.Read(r, binary.BigEndian, &f.Header)
	if err != nil {
		return
	}

	f.Blob = make([]byte, f.Length)
	_, err = io.ReadAtLeast(r, f.Blob, len(f.Blob))

	return
}
