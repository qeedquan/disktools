package fat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	stdpath "path"
	"sort"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/qeedquan/disktools/iod"
)

var (
	ErrNotDir = errors.New("not a directory")
)

const (
	fatDirsz = 32
)

type FileSystem struct {
	rw  iod.RW
	opt *FileSystemOptions
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
	label       string
	fstype      string

	rootdir File
}

type FileSystemOptions struct {
	Case bool
}

type File struct {
	name     string
	dir      Dir
	sect     int64
	dirstart int64
	dirpos   int64
	direof   bool
	off      int64
	fs       *FileSystem
}

func (f *File) Name() string     { return f.name }
func (f *File) IsDir() bool      { return f.dir.Attr&DIRECTORY != 0 && f.dir.Attr != 0xF }
func (f *File) Size() int64      { return int64(f.dir.Length) }
func (f *File) Sys() interface{} { return f }
func (f *File) Close() error     { return nil }

func (f *File) ModTime() time.Time {
	return time.Time{}
}

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

func (f *File) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.dirpos = f.dirstart
		f.direof = false
	case io.SeekCurrent:
		off += f.off
	case io.SeekEnd:
		off = int64(f.dir.Length) + off
	default:
		return 0, os.ErrInvalid
	}

	if off < 0 {
		return 0, os.ErrInvalid
	}

	f.off = off
	return off, nil
}

func (f *File) Readdir(n int) ([]os.FileInfo, error) {
	if !f.IsDir() {
		return nil, &os.PathError{"readdir", f.name, ErrNotDir}
	}
	if f.direof {
		return nil, io.EOF
	}

	var lfns []LFN
	var dir Dir
	var buf [fatDirsz]byte
	var fis []os.FileInfo

	dp := f.dirpos
	sr := io.NewSectionReader(f.fs.rw, dp, math.MaxUint32)
	for i := 0; n <= 0 || i < n; dp += fatDirsz {
		err := binary.Read(sr, binary.LittleEndian, &buf)
		if err != nil {
			return fis, err
		}

		if buf[0] == 0 {
			f.direof = true
			return fis, nil
		}

		bp := bytes.NewReader(buf[:])
		switch {
		case buf[11]&0xf == 0xf: // lfn
			var lfn LFN
			binary.Read(bp, binary.LittleEndian, &lfn)
			if lfn.Cluster == 0 && (lfn.Name1[0] != 0 || lfn.Name1[1] != 0) {
				lfns = append(lfns, lfn)
			}

		case buf[11]&0x2 != 0: // hidden
			lfns = lfns[:0]

		case buf[0] == 0xe5: // deleted
			lfns = lfns[:0]

		default:
			binary.Read(bp, binary.LittleEndian, &dir)

			var name string
			if len(lfns) > 0 {
				name = lfnName(lfns)
				lfns = lfns[:0]
			} else {
				name = string(dir.Name[:])
				name = strings.TrimRight(name, " ")

				ext := string(dir.Ext[:])
				ext = strings.TrimRight(ext, " ")
				if ext != "" {
					name += "." + ext
				}

				name = strings.ToLower(name)
			}

			fis = append(fis, &File{
				name:   name,
				sect:   dp,
				dirpos: dp,
				dir:    dir,
			})

			i++
		}

		f.dirpos = dp
	}
	return fis, nil
}

func NewFileSystem(rw iod.RW, opt *FileSystemOptions) (*FileSystem, error) {
	if opt == nil {
		opt = &FileSystemOptions{}
	}
	fs := &FileSystem{
		rw:  rw,
		cwd: "/",
		opt: opt,
	}

	var pbs PBS
	var pbs32 PBS32

	sr := io.NewSectionReader(rw, 0, math.MaxUint32)
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
		fs.label = strings.TrimRight(string(pbs.Label[:]), "\x00")
		fs.fstype = strings.TrimRight(string(pbs.Fstype[:]), "\x00")
	} else {
		fs.rootaddr = fs.fataddr + fs.nfats*fs.fatsz
		i := fs.rootsz*fatDirsz + fs.sectsz - 1
		i /= fs.sectsz
		fs.dataaddr = fs.rootaddr + i
		fs.label = strings.TrimRight(string(pbs.Label[:]), "\x00")
		fs.fstype = strings.TrimRight(string(pbs.Fstype[:]), "\x00")
	}
	fs.fatclusters = fs.nresrv + (fs.volsz-fs.dataaddr)/fs.clustsz

	if fs.fatbits != 32 {
		if fs.fatclusters < 4087 {
			fs.fatbits = 12
		} else {
			fs.fatbits = 16
		}
	}
	fs.rootdir = File{
		name:     "/",
		sect:     fs.rootaddr * fs.sectsz,
		dirstart: fs.rootaddr * fs.sectsz,
		dirpos:   fs.rootaddr * fs.sectsz,
		dir: Dir{
			Name: [8]byte{'/'},
			Attr: DIRECTORY,
		},
		fs: fs,
	}

	return fs, nil
}

func (fs *FileSystem) Getwd() (string, error) {
	return fs.cwd, nil
}

func (fs *FileSystem) Chmod(name string, mode os.FileMode) error {
	f, err := fs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if mode&0444 == 0444 {
		f.dir.Attr |= RDONLY
	} else {
		f.dir.Attr &^= RDONLY
	}
	return nil
}

func (fs *FileSystem) Chdir(dir string) error {
	f, err := fs.Open(dir)
	if err != nil {
		return err
	}

	if !f.IsDir() {
		return &os.PathError{"chdir", dir, ErrNotDir}
	}
	fs.cwd = stdpath.Clean(dir)

	return nil
}

func (fs *FileSystem) Open(name string) (*File, error) {
	if stdpath.IsAbs(name) {
		name = stdpath.Clean(name)
	} else {
		name = stdpath.Join(fs.cwd, name)
	}

	f := &File{}
	*f = fs.rootdir
	if name == "/" {
		return f, nil
	}

	p := splitPath(name)
	for i := len(p) - 1; i >= 0; i-- {
	loop:
		for {
			fi, err := f.Readdir(1024)
			if err == io.EOF {
				return nil, &os.PathError{"open", name, os.ErrNotExist}
			}
			if err != nil {
				return nil, err
			}

			for _, fi := range fi {
				if fs.compareName(fi.Name(), p[i]) == 0 {
					f = fi.(*File)
					continue loop
				}
			}
		}
	}

	return f, nil
}

func (fs *FileSystem) compareName(a, b string) int {
	if !fs.opt.Case {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}
	return strings.Compare(a, b)
}

func lfnName(lfns []LFN) string {
	sort.Slice(lfns, func(i, j int) bool {
		return lfns[i].Seq < lfns[j].Seq
	})

	var name string
	for _, l := range lfns {
		r1 := utf16.Decode(l.Name0[:])
		r2 := utf16.Decode(l.Name1[:])
		r3 := utf16.Decode(l.Name2[:])
		name += string(r1) + string(r2) + string(r3)
	}

	i := strings.Index(name, "\x00")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func splitPath(path string) []string {
	var toks []string
	for str := path; str != ""; {
		dir, base := stdpath.Split(str)
		if dir == "" && base == "" {
			break
		}

		if len(dir) > 0 && dir[len(dir)-1] == '/' {
			dir = dir[:len(dir)-1]
		}

		if base == "" {
			if dir == "" {
				dir = "."
			}
			toks = append(toks, dir)
			break
		}

		toks = append(toks, base)
		str = dir
	}
	return toks
}

func (fs *FileSystem) String() string {
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "Type:           FAT%d\n", fs.fatbits)
	if fs.label != "" {
		fmt.Fprintf(b, "Label:          %s\n", fs.label)
	}
	if fs.fstype != "" {
		fmt.Fprintf(b, "FS Type:        %s\n", fs.fstype)
	}
	fmt.Fprintf(b, "Sector size:    %d\n", fs.sectsz)
	fmt.Fprintf(b, "Cluster size:   %d\n", fs.clustsz)
	fmt.Fprintf(b, "Reserved:       %d\n", fs.nresrv)
	fmt.Fprintf(b, "Number of FATs: %d\n", fs.nfats)
	fmt.Fprintf(b, "Root size:      %d\n", fs.rootsz)
	fmt.Fprintf(b, "Volume size:    %d\n", fs.volsz)
	fmt.Fprintf(b, "FAT address:    %d\n", fs.fataddr)
	fmt.Fprintf(b, "FAT size:       %d\n", fs.fatsz)
	fmt.Fprintf(b, "Data address:   %d\n", fs.dataaddr)
	fmt.Fprintf(b, "Root address:   %d\n", fs.rootaddr)
	fmt.Fprintf(b, "Root offset:    %d\n", fs.rootstart)
	fmt.Fprintf(b, "FAT cluster:    %d", fs.fatclusters)
	return b.String()
}
