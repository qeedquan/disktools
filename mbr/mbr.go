package mbr

import (
	"errors"
	"io"

	"github.com/qeedquan/disktools/endian"
)

type Record struct {
	Part [4]Part
}

type Part struct {
	Bootable uint8
	Start    CHS
	Type     uint8
	End      CHS
	LBA      uint32
	Sectors  uint32
}

type CHS struct {
	Head     uint64
	Sector   uint64
	Cylinder uint64
}

var (
	ErrHeader = errors.New("mbr: invalid boot signature")
)

func Open(r io.ReaderAt) (*Record, error) {
	var mbr [512]byte
	_, err := r.ReadAt(mbr[:], 0)
	if err != nil {
		return nil, err
	}

	if mbr[0x1fe] != 0x55 || mbr[0x1ff] != 0xaa {
		return nil, ErrHeader
	}

	var part [4]Part
	for i := range part {
		part[i] = readPart(mbr[0x1be+i*16:])
	}

	return &Record{
		Part: part,
	}, nil
}

func readPart(b []byte) Part {
	return Part{
		Bootable: b[0],
		Start: CHS{
			Head:     uint64(b[1]),
			Sector:   uint64(b[2]) & 0xbf,
			Cylinder: uint64(b[2])<<8 | uint64(b[3]),
		},
		Type: b[4],
		End: CHS{
			Head:     uint64(b[5]),
			Sector:   uint64(b[6]) & 0xbf,
			Cylinder: uint64(b[6])<<8 | uint64(b[7]),
		},
		LBA:     endian.Read32le(b[8:]),
		Sectors: endian.Read32le(b[12:]),
	}
}
