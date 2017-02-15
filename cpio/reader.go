package cpio

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

var (
	ErrHeader  = errors.New("cpio: invalid header")
	ErrArchive = errors.New("cpio: invalid archive")
)

type hdrbin struct {
	Magic  uint16
	Dev    uint16
	Ino    uint16
	Mode   uint16
	UID    uint16
	GID    uint16
	Nlink  uint16
	Rdev   uint16
	Mtime  uint64
	Namesz uint16
	Filesz uint64
}

type hdrodc struct {
	Magic  [6]byte
	Dev    [6]byte
	Ino    [6]byte
	Mode   [6]byte
	UID    [6]byte
	GID    [6]byte
	Nlink  [6]byte
	Rdev   [6]byte
	Mtime  [11]byte
	Namesz [6]byte
	Filesz [11]byte
}

type hdrnewc struct {
	Magic     [6]byte
	Ino       [8]byte
	Mode      [8]byte
	UID       [8]byte
	GID       [8]byte
	Nlink     [8]byte
	Mtime     [8]byte
	Filesz    [8]byte
	Devmajor  [8]byte
	Devminor  [8]byte
	Rdevmajor [8]byte
	Rdevminor [8]byte
	Namesz    [8]byte
	Check     [8]byte
}

type Header struct {
	Name  string
	UID   string
	GID   string
	Size  int64
	Mtime time.Time
}

type Reader struct {
	r     io.Reader
	b     *bufio.Reader
	nleft int64
	pad   int64
}

const (
	hdrbinsz  = 26
	hdrodcsz  = 76
	hdrnewcsz = 110
)

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
		b: bufio.NewReader(r),
	}
}

func (cr *Reader) skip(n int64) {
	for n > 0 {
		m := int64(math.MaxInt32)
		if n < m {
			m = n
		}
		cr.b.Discard(int(m))
		n -= m
	}
}

func (cr *Reader) Next() (*Header, error) {
	if cr.nleft > 0 || cr.pad > 0 {
		cr.skip(cr.nleft)
		cr.skip(cr.pad)
		cr.nleft, cr.pad = 0, 0
	}

	magic, err := cr.b.Peek(6)
	if err != nil {
		return nil, wrapError(err)
	}

	var (
		order binary.ByteOrder = binary.LittleEndian
		hp    hdrnewc
		stsz  int64
	)

	switch {
	case uint16(magic[1])|uint16(magic[0])<<8 == 070707:
		order = binary.BigEndian
		fallthrough
	case uint16(magic[0])|uint16(magic[1])<<8 == 070707:
		var h hdrbin
		err = binary.Read(cr.b, order, &h)
		copyStructField(&hp, &h)
		stsz = hdrbinsz

	case string(magic[:]) == "070707":
		var h hdrodc
		err = binary.Read(cr.b, order, &h)
		copyStructField(&hp, &h)
		stsz = hdrodcsz

	case string(magic[:]) == "070701":
		fallthrough
	case string(magic[:]) == "070702":
		err = binary.Read(cr.b, order, &hp)
		stsz = hdrnewcsz

	default:
		return nil, ErrHeader
	}
	if err != nil {
		return nil, wrapError(err)
	}

	namesz, _ := strconv.ParseInt(string(hp.Namesz[:]), 16, 64)
	filesz, _ := strconv.ParseInt(string(hp.Filesz[:]), 16, 64)
	if namesz <= 0 || filesz < 0 {
		return nil, ErrArchive
	}

	name := make([]byte, namesz)
	_, err = io.ReadAtLeast(cr.b, name, len(name))
	if err != nil {
		return nil, wrapError(err)
	}

	pad := (stsz+namesz+3)&^0x3 - (stsz + namesz)
	if pad > 0 {
		cr.b.Discard(int(pad))
	}

	cr.nleft = filesz
	cr.pad = (filesz+3)&^0x3 - filesz
	hdr := &Header{
		Name: strings.TrimRight(string(name), "\x00"),
		UID:  strings.TrimRight(string(hp.UID[:]), "\x00"),
		GID:  strings.TrimRight(string(hp.GID[:]), "\x00"),
		Size: filesz,
	}

	if hdr.Name == "TRAILER!!!" {
		return nil, io.EOF
	}
	return hdr, nil
}

func (cr *Reader) Read(b []byte) (int, error) {
	if cr.nleft == 0 {
		return 0, io.EOF
	}

	nr := cr.nleft
	if int64(len(b)) < nr {
		nr = int64(len(b))
	}

	n, err := cr.b.Read(b[:nr])
	cr.nleft -= int64(n)
	return n, err
}
