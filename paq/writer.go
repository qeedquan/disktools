package paq

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/adler32"
	"io"
	"io/ioutil"
	"os"
	"os/user"
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
	offset int64
	qid    uint32
}

type WriteOptions struct {
	Time      time.Time
	Compress  bool
	Label     string
	Version   int
	BlockSize int
	Uid       string
	Gid       string
}

func DefaultWriteOptions() *WriteOptions {
	uid, gid := "glenda", "glenda"
	u, err := user.Current()
	if err == nil {
		uid = u.Uid
		gid = u.Gid
	}

	o := &WriteOptions{
		Time:      time.Now(),
		Compress:  true,
		BlockSize: 4096,
		Version:   Version,
		Uid:       uid,
		Gid:       gid,
	}
	return o
}

func NewWriter(w io.Writer, o *WriteOptions) (*Writer, error) {
	if o == nil {
		o = DefaultWriteOptions()
	}

	if !(MinBlockSize <= o.BlockSize && o.BlockSize <= MaxBlockSize) {
		return nil, fmt.Errorf("paq: invalid block size")
	}

	p := &Writer{
		o:   o,
		w:   bufio.NewWriter(w),
		dg:  sha1.New(),
		qid: 1,
	}

	return p, nil
}

func (w *Writer) WriteHeader() {
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
	copy(b[12:], w.o.Label)

	w.write(b[:])
}

func (w *Writer) WriteDir(dir string, di os.FileInfo) (*Dir, error) {
	fis, err := ioutil.ReadDir(di.Name())
	if err != nil {
		return nil, fmt.Errorf("paq: %v", err)
	}

	var (
		pd     *Dir
		n, nb  int
		offset int64
		b      = make([]byte, w.o.BlockSize)
		p      = make([]byte, w.o.BlockSize)
	)
	for _, fi := range fis {
		name := filepath.Join(dir, fi.Name())
		if fi.IsDir() {
			pd, err = w.WriteDir(dir, fi)
		} else {
			fd, err := os.Open(name)
			if err != nil {
				return nil, err
			}
			pd, err = w.WriteFile(fd, fi)
			fd.Close()
		}

		if err != nil {
			return nil, fmt.Errorf("paq: ", err)
		}

		if n+dirsize(pd) >= len(b) {
			for i := n; i < len(b); i++ {
				b[i] = 0
			}
			offset = w.WriteBlock(b, DirBlock)
			n = 0
			if nb >= len(b)/offsetSize {
				return nil, fmt.Errorf("paq: directory too big for block size: %s", dir)
			}
			put4(p[nb*offsetSize:], uint32(offset))
			nb++
		}

		if n+dirsize(pd) >= len(b) {
			return nil, fmt.Errorf("paq: directory too big for block size: %s", dir)
		}

		putdir(b[n:], pd)
		n += dirsize(pd)
		pd = nil
	}

	if n > 0 {
		for i := n; i < len(b); i++ {
			b[i] = 0
		}
		offset = w.WriteBlock(b, DirBlock)
		if nb >= len(b)/offsetSize {
			return nil, fmt.Errorf("paq: directory too big for block size: %s", dir)
		}
		put4(p[nb*offsetSize:], uint32(offset))
	}

	offset = w.WriteBlock(p, PointerBlock)
	d := w.allocDir(di, offset)
	return d, nil
}

func (w *Writer) WriteFile(r io.Reader, fi os.FileInfo) (*Dir, error) {
	b := make([]byte, w.o.BlockSize)
	p := make([]byte, w.o.BlockSize)

	nb := 0
	n := 0
	tot := 0
	for {
		nn, err := r.Read(b[n:])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("paq: failed to read %s: %v", fi.Name(), err)
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
			return nil, fmt.Errorf("paq: file too big for block size")
		}

		offset := w.WriteBlock(b, DataBlock)
		put4(p[nb*offsetSize:], uint32(offset))
		nb++
		n = 0
	}
	offset := w.WriteBlock(p, PointerBlock)

	d := w.allocDir(fi, offset)
	d.Length = uint32(tot)

	return d, nil
}

func (w *Writer) allocDir(fi os.FileInfo, offset int64) *Dir {
	mode := fi.Mode() & 0777
	if fi.IsDir() {
		mode |= dmdir
	}

	mtime := fi.ModTime().Unix()

	d := &Dir{
		Qid:    w.qid,
		Name:   fi.Name(),
		Length: uint32(fi.Size()),
		Mode:   uint32(mode),
		Uid:    w.o.Uid,
		Gid:    w.o.Gid,
		Offset: uint32(offset),
		Mtime:  uint32(mtime),
	}
	w.qid++
	return d
}

func (w *Writer) WriteTrailer(root uint32) {
	var b [TrailerSize]byte
	put4(b[:], TrailerMagic)
	put4(b[4:], root)

	w.dg.Write(b[:8])
	d := w.dg.Sum(nil)
	copy(b[8:], d)

	w.write(b[:])
}

func (w *Writer) write(b []byte) {
	n, _ := w.w.Write(b)
	w.dg.Write(b)
	w.offset += int64(n)
}

func (w *Writer) WriteBlock(b []byte, typ int) int64 {
	off := w.offset

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
			bh.Encoding = DeflateEnc
			bh.Size = uint32(len(b))
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

func (w *Writer) WriteBlockDir(d *Dir) int64 {
	b := make([]byte, w.o.BlockSize)
	putdir(b, d)
	return w.WriteBlock(b, DirBlock)
}

func (w *Writer) Close() error {
	return w.w.Flush()
}

func putdir(b []byte, d *Dir) {
	put2(b[:], uint16(dirsize(d)))
	put4(b[2:], d.Qid)
	put4(b[6:], d.Mode)
	put4(b[10:], d.Mtime)
	put4(b[14:], d.Length)
	put4(b[18:], d.Offset)

	n := 22
	n += puts(b[n:], d.Name)
	n += puts(b[n:], d.Uid)
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

func puts(b []byte, s string) int {
	n := uint16(len(s))
	put2(b, n+2)
	copy(b[2:], s[:n])
	return int(n) + 2
}
