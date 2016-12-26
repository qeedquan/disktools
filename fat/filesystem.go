package fat

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/qeedquan/disktools/iod"
)

type FileSystem struct {
	rw  iod.RW
	cwd string

	sectsz  int
	clustsz int
	resrv   int
	nfats   int
	fatsz   int
	rootsz  int
	volsz   int
}

type File struct {
}

func NewFileSystem(rw iod.RW) (*FileSystem, error) {
	fs := &FileSystem{
		rw:  rw,
		cwd: "/",
	}

	var pbs PBS
	var pbs32 PBS32

	sr := io.NewSectionReader(rw, 0, math.MaxInt64)
	err := binary.Read(sr, binary.LittleEndian, &pbs)
	if err == nil && pbs.Fatsz == 0 {
		sr.Seek(0, io.SeekStart)
		err = binary.Read(sr, binary.LittleEndian, &pbs32)
	}

	if err != nil {
		return nil, err
	}

	fs.sectsz = int(pbs.Sectsz)
	fs.clustsz = int(pbs.Clustsz)
	fs.resrv = int(pbs.Resrv)
	fs.nfats = int(pbs.NumFats)
	fs.rootsz = int(pbs.Rootsz)
	fs.volsz = int(pbs.Volsz)
	if fs.volsz == 0 {
		fs.volsz = int(pbs.Bigvolsz)
	}
	fs.fatsz = int(pbs.Fatsz)

	return fs, nil
}

func (fs *FileSystem) Getwd() (string, error) {
	return fs.cwd, nil
}

func (fs *FileSystem) Open(name string) (*File, error) {
	return nil, nil
}
