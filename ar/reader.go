package ar

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"time"
)

const magic = "!<arch>\n"

var (
	ErrHeader = errors.New("ar: invalid header")
)

type hdro struct {
	Name    [16]byte
	Mtime   [12]byte
	UID     [6]byte
	GID     [6]byte
	Mode    [8]byte
	Size    [10]byte
	Trailer [2]byte
}

type Header struct {
	Name  string
	UID   string
	GID   string
	Mode  os.FileMode
	Size  uint64
	Mtime time.Time
}

type Reader struct {
	r    io.Reader
	b    *bufio.Reader
	left uint64
	pad  uint64
}

func NewReader(r io.Reader) (*Reader, error) {
	var sig [len(magic)]byte
	_, err := io.ReadAtLeast(r, sig[:], len(sig))
	if err != nil {
		return nil, wk(err)
	}

	if string(sig[:]) != magic {
		return nil, ErrHeader
	}

	return &Reader{
		r: r,
		b: bufio.NewReader(r),
	}, nil
}

func (cr *Reader) skip(n uint64) {
	for n > 0 {
		m := uint64(math.MaxInt32)
		if n < m {
			m = n
		}
		cr.b.Discard(int(m))
		n -= m
	}
}

func (cr *Reader) Next() (*Header, error) {
	if cr.left > 0 || cr.pad > 0 {
		cr.skip(cr.left)
		cr.skip(cr.pad)
		cr.left, cr.pad = 0, 0
	}

	var h hdro
	err := binary.Read(cr.b, binary.LittleEndian, &h)
	if err != nil {
		return nil, wk(err)
	}

	cr.left, err = strconv.ParseUint(trim(h.Size[:]), 0, 64)
	if err != nil {
		return nil, wk(err)
	}
	cr.pad = cr.left & 1

	mode, _ := strconv.ParseInt(trim(h.Mode[:]), 8, 64)
	mtime, _ := strconv.ParseInt(trim(h.Mtime[:]), 0, 64)

	return &Header{
		Name:  trim(h.Name[:]),
		UID:   trim(h.UID[:]),
		GID:   trim(h.GID[:]),
		Mode:  os.FileMode(mode),
		Size:  cr.left,
		Mtime: time.Unix(mtime, 0),
	}, nil
}

func (cr *Reader) Read(b []byte) (int, error) {
	if cr.left == 0 {
		return 0, io.EOF
	}

	nr := cr.left
	if uint64(len(b)) < nr {
		nr = uint64(len(b))
	}

	n, err := cr.b.Read(b[:nr])
	cr.left -= uint64(n)
	return n, err
}
