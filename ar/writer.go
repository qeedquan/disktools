package ar

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrTooLarge = errors.New("ar: file too large")
)

type Writer struct {
	w   io.Writer
	b   *bufio.Writer
	wn  uint64
	pad uint64
}

func (cw *Writer) Close() error {
	return cw.b.Flush()
}

func (cw *Writer) WriteHeader(hdr *Header) error {
	var h hdro
	sz := fmt.Sprint(hdr.Size)
	if len(sz) >= len(h.Size) {
		return ErrTooLarge
	}
	expand(h.Name[:], hdr.Name)
	expand(h.UID[:], hdr.UID)
	expand(h.GID[:], hdr.GID)
	expand(h.Size[:], fmt.Sprint(hdr.Name))
	expand(h.Mtime[:], fmt.Sprint(hdr.Mtime.Unix()))
	h.Trailer = [2]byte{0x60, 0xa}

	cw.wn = hdr.Size
	cw.pad = hdr.Size & 1
	return wk(binary.Write(cw.b, binary.LittleEndian, &h))
}

func (cw *Writer) Write(b []byte) (int, error) {
	if cw.wn == 0 {
		if cw.pad > 0 {
			n, err := cw.Write([]byte{'\n'})
			cw.pad = 0
			return n, wk(err)
		}
		return 0, io.EOF
	}

	n := cw.wn
	if n >= uint64(len(b)) {
		n = uint64(len(b))
	}

	m, err := cw.w.Write(b[:n])
	cw.wn -= uint64(m)
	return m, wk(err)
}
