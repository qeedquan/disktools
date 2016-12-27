package fat

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path"

	"github.com/qeedquan/disktools/iod"
)

const (
	fatDirsz = 32
)

type FileSystem struct {
	rw  iod.RW
	cwd string

	sectsz      int64
	clustsz     int64
	nresrv      int64
	nfats       int64
	rootsz      int64
	volsz       int64
	mediadesc   int64
	fatsz       int64
	fataddr     int64
	fatbits     int64
	dataaddr    int64
	rootaddr    int64
	rootstart   int64
	fatclusters int64
}

type File struct {
	name string
	dir  Dir
}

func (f *File) Name() string     { return f.name }
func (f *File) IsDir() bool      { return f.dir.Attr&DIRECTORY != 0 && f.dir.Attr != 0xF }
func (f *File) Size() int64      { return int64(f.dir.Length) }
func (f *File) Sys() interface{} { return f }

func (f *File) Mode() os.FileMode {
	m := os.FileMode(0777)
	v := f.dir.Attr
	if v&DIRECTORY != 0 {
		m |= os.ModeDir
	}

	if v&DEVICE != 0 {
		m |= os.ModeDevice
	}

	if v&RDONLY != 0 {
		m &^= 0333
	}

	return m
}

func NewFileSystem(rw iod.RW) (*FileSystem, error) {
	fs := &FileSystem{
		rw:  rw,
		cwd: "/",
	}

	var pbs PBS
	var pbs32 PBS32

	sr := io.NewSectionReader(rw, 0, math.MaxInt32)
	err := binary.Read(sr, binary.LittleEndian, &pbs)
	if err == nil && pbs.Fatsz == 0 {
		sr.Seek(0, io.SeekStart)
		err = binary.Read(sr, binary.LittleEndian, &pbs32)
	}

	if err != nil {
		return nil, err
	}

	fs.sectsz = int64(pbs.Sectsz)
	fs.clustsz = int64(pbs.Clustsz)
	fs.nresrv = int64(pbs.Resrv)
	fs.nfats = int64(pbs.NumFats)
	fs.rootsz = int64(pbs.Rootsz)
	fs.volsz = int64(pbs.Volsz)
	if fs.volsz == 0 {
		fs.volsz = int64(pbs.Bigvolsz)
	}
	fs.mediadesc = int64(pbs.Mediadesc)
	fs.fatsz = int64(pbs.Fatsz)
	fs.fataddr = int64(pbs.Resrv)

	if fs.fatsz == 0 {
		fs.fatbits = 32
		fs.fatsz = int64(pbs32.Fatsz32)
		if fs.fatsz == 0 {
			return nil, fmt.Errorf("fat size is 0")
		}
		fs.dataaddr = fs.fataddr + fs.nfats*fs.fatsz
		fs.rootstart = int64(pbs32.Rootstart)

		if ext := pbs32.Extflags; ext&0x80 != 0 {
			for i := uint(0); i < 4; i++ {
				if ext&(1<<i) != 0 {
					fs.fataddr += int64(i) * fs.fatsz
					fs.nfats = 1
					break
				}
			}
		}
	} else {
		fs.rootaddr = fs.fataddr + fs.nfats*fs.fatsz
		i := fs.rootsz*fatDirsz + fs.sectsz - 1
		i /= fs.sectsz
		fs.dataaddr = fs.rootaddr + i
	}
	fs.fatclusters = fs.nresrv + (fs.volsz-fs.dataaddr)/fs.clustsz

	if fs.fatbits != 32 {
		if fs.fatclusters < 4087 {
			fs.fatbits = 12
		} else {
			fs.fatbits = 16
		}
	}

	fs.Open("/")

	return fs, nil
}

func (fs *FileSystem) Getwd() (string, error) {
	return fs.cwd, nil
}

func (fs *FileSystem) Chdir(dir string) error {
	f, err := fs.Open(dir)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return &os.PathError{"chdir", dir, fmt.Errorf("is not a directory")}
	}
	fs.cwd = path.Clean(dir)

	return nil
}

func (fs *FileSystem) Open(name string) (*File, error) {
	if path.IsAbs(name) {
		name = path.Clean(name)
	} else {
		name = path.Join(fs.cwd, name)
	}

	var dir Dir
	sect := fs.rootaddr * fs.sectsz
	sr := io.NewSectionReader(fs.rw, sect, math.MaxInt32)
	for i := 0; i < 32; i++ {
		err := binary.Read(sr, binary.LittleEndian, &dir)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%#v\n", dir)
	}

	return nil, nil
}
