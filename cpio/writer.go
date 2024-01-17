package cpio

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

var (
	ErrTooLarge = errors.New("cpio: file too large")
)

type Writer struct {
	w  io.Writer
	b  *bufio.Writer
	wn int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
		b: bufio.NewWriter(w),
	}
}

func (cw *Writer) Close() error {
	cw.flushFile()
	cw.writeTrailer()
	return wrapError(cw.b.Flush())
}

func (cw *Writer) flushFile() {
	if cw.wn <= 0 {
		return
	}

	var zeroes [4]byte
	pad := (cw.wn+3)&^3 - cw.wn
	cw.b.Write(zeroes[:pad])
	cw.wn = 0
}

func (cw *Writer) Write(b []byte) (int, error) {
	if cw.wn > math.MaxUint32 {
		return 0, ErrTooLarge
	}

	if cw.wn-math.MaxInt32 < len(b) {
		b = b[:cw.wn-math.MaxInt32]
	}

	n, err := cw.b.Write(b)
	cw.wn += n
	return n, wrapError(err)
}

func (cw *Writer) WriteHeader(hdr *Header) error {
	cw.flushFile()

	h := hdrnewc{
		Magic: [6]byte{'0', '7', '0', '7', '0', '1'},
	}
	namesz := fmt.Sprintf("%X", len(hdr.Name)+1)
	filesz := fmt.Sprintf("%X", hdr.Size)
	copy(h.Namesz[:], namesz)
	copy(h.Filesz[:], filesz)

	binary.Write(cw.b, binary.LittleEndian, &h)
	cw.b.Write([]byte(hdr.Name))
	cw.b.Write([]byte{0x00})

	var zeroes [4]byte
	pad := (hdrnewcsz+len(hdr.Name)+1+3)&^0x3 - (hdrnewcsz + len(hdr.Name) + 1)
	cw.b.Write(zeroes[:pad])

	return nil
}

func (cw *Writer) writeTrailer() {
	h := hdrnewc{
		Magic:  [6]byte{'0', '7', '0', '7', '0', '1'},
		Namesz: [8]byte{'0', '0', '0', '0', '0', '0', '0', 'B'},
	}
	binary.Write(cw.b, binary.LittleEndian, &h)
	cw.b.Write([]byte("TRAILER!!!\x00"))
	cw.b.Write([]byte{0, 0, 0})
}
