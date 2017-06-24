package paq

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/adler32"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	offsetSize = 4
)

type Writer struct {
	o      *WriteOptions
	w      *bufio.Writer
	dg     hash.Hash
	off    int64
	pd     *Dir
	closed bool
}

type WriteOptions struct {
	Time      time.Time
	Compress  bool
	Label     string
	Version   int
	BlockSize int
}

func NewWriter(w io.Writer, o *WriteOptions) (*Writer, error) {
	if o == nil {
		o = &WriteOptions{
			Time:      time.Now(),
			Compress:  true,
			BlockSize: 4096,
			Version:   Version,
		}
	}

	if !(MinBlockSize <= o.BlockSize && o.BlockSize <= MaxBlockSize) {
		return nil, fmt.Errorf("paq: invalid block size")
	}

	p := &Writer{
		o: o,
		w: bufio.NewWriter(w),
	}
	p.writeHeader()

	return p, nil
}

func (w *Writer) WriteDir(dir string, di os.FileInfo) error {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return &os.PathError{Op: "readdir", Path: dir, Err: err}
	}

	for _, fi := range fis {
		name := filepath.Join(dir, fi.Name())
		if fi.IsDir() {
			err = w.WriteDir(name, fi)
		} else {
			fd, err := os.Open(name)
			if err != nil {
				return err
			}
			err = w.WriteFile(name, fd)
			fd.Close()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) WriteFile(name string, r io.Reader) error {
	b := make([]byte, w.o.BlockSize)
	p := make([]byte, w.o.BlockSize)

	nb := 0
	n := 0
	tot := 0
	for {
		nn, err := r.Read(b[n:])
		if err != nil && err != io.EOF {
			return &os.PathError{Op: "read", Path: name, Err: err}
		}
		tot += nn
		if err == io.EOF {
			if n == 0 {
				break
			}
			// pad out last block
			for i := n; i < len(b); i++ {
				b[i] = 0
			}
			nn = len(b) - n
		}
		n += nn
		if n < len(b) {
			continue
		}
		if nb >= len(b)/offsetSize {
			return fmt.Errorf("file too big for block size")
		}

		off := w.writeBlock(b, DataBlock)
		put4(p[nb*offsetSize:], uint32(off))
		nb++
		n = 0
	}
	w.writeBlock(p, PointerBlock)

	return nil
}

func (w *Writer) Close() error {
	if w.closed {
		return nil
	}

	defer func() {
		w.closed = true
	}()

	if w.pd != nil {
		return nil
	}
	w.writeTrailer(0)
	return w.w.Flush()
}

func (w *Writer) write(b []byte) {
	n, _ := w.w.Write(b)
	w.dg.Write(b)
	w.off += int64(n)
}

func (w *Writer) writeHeader() {
	var b [HeaderSize]byte
	if w.o.BlockSize < 65536 {
		put4(b[:], HeaderMagic)
		put2(b[4:], uint16(w.o.Version))
		put2(b[6:], uint16(w.o.BlockSize))
	} else {
		put4(b[:], BigHeaderMagic)
		put2(b[2:], uint16(w.o.Version))
		put4(b[4:], uint32(w.o.BlockSize))
	}
	put4(b[8:], uint32(w.o.Time.Unix()))
	puts(b[12:], w.o.Label)

	w.write(b[:])
}

func (w *Writer) writeBlock(b []byte, typ int) int64 {
	off := w.off

	bh := Block{
		Magic:    BlockMagic,
		Size:     uint32(w.o.BlockSize),
		Type:     uint8(typ),
		Encoding: NoEnc,
		Adler32:  adler32.Checksum(b),
	}

	if w.o.Compress {
		p := new(bytes.Buffer)
		f, _ := flate.NewWriter(p, 6)
		_, err := f.Write(b)
		xerr := f.Close()
		if err == nil && xerr == nil {
			b = p.Bytes()
		}
	}

	var bp [BlockSize]byte
	if bh.Size < 65536 {
		put4(bp[:], bh.Magic)
		put2(bp[4:], uint16(bh.Size))
	} else {
		put2(bp[:], BigBlockMagic)
		put4(bp[2:], bh.Size)
	}
	bp[6] = bh.Type
	bp[7] = bh.Encoding
	put4(bp[8:], bh.Adler32)
	w.write(bp[:])
	w.write(b)

	return off
}

func (w *Writer) writeTrailer(root uint32) {
	var b [TrailerSize]byte
	put4(b[:], TrailerMagic)
	put4(b[4:], root)
	w.dg.Write(b[:8])

	d := w.dg.Sum(nil)
	copy(b[8:], d)

	w.write(b[:])
}

func putdir(b []byte, d *Dir) {
	put2(b[:], uint16(dirsize(d)))
	put4(b[2:], d.Qid)
	put4(b[6:], d.Mode)
	put4(b[10:], d.Mtime)
	put4(b[14:], d.Length)
	put4(b[18:], d.Offset)

	n := 22
	puts(b[n:], d.Name)
	n += len(d.Name)
	puts(b[n:], d.Uid)
	n += len(d.Uid)
	puts(b[n:], d.Gid)
}

func dirsize(d *Dir) int {
	return MinDirSize + len(d.Name) + len(d.Uid) + len(d.Gid)
}

func put4(b []byte, v uint32) {
	binary.BigEndian.PutUint32(b, v)
}

func put2(b []byte, v uint16) {
	binary.BigEndian.PutUint16(b, v)
}

func puts(b []byte, s string) {
	n := uint16(len(s))
	put2(b, n+2)
	copy(b[2:], s[:n])
}