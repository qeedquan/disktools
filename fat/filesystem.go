package fat

import "github.com/qeedquan/disktools/iod"

type FileSystem struct {
	rw  iod.RW
	cwd string
}

type File struct {
}

func NewFileSystem(rw iod.RW) *FileSystem {
	return &FileSystem{
		rw: rw,
	}
}

func (fs *FileSystem) Open(name string) (*File, error) {
	return nil, nil
}
