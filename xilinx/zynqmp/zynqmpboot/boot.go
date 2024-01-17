package zynqmpboot

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

type Header struct {
	Vectors         [8]uint32
	Width           uint32
	Sig             [4]byte
	KeySource       uint32
	FSBLEntry       uint32
	SourceOffset    uint32
	PMUImageSize    uint32
	PMUTotalSize    uint32
	FSBLImageSize   uint32
	FSBLTotalSize   uint32
	FSBLImageFlags  uint32
	Checksum        uint32
	BlackKey        [32]byte
	ShutterValue    uint32
	UserDefined     [40]byte
	ImageOffset     uint32
	PartitionOffset uint32
	SecureIV        [12]byte
	BlackKeyIV      [12]byte
}

type ImageHeaderTable struct {
	Version             uint32
	NumImages           uint32
	PartitionOffset     uint32
	ImageOffset         uint32
	AuthOffset          uint32
	SecondaryBootDevice uint32
	Padding             [32]byte
	Checksum            uint32
}

type ImageHeader struct {
	Next          uint32
	Partition     uint32
	Reserved      uint32
	NumPartitions uint32
}

type Image struct {
	ImageHeader
	Name string
}

type PartitionHeader struct {
	EncryptedSize       uint32
	UnencryptedSize     uint32
	TotalSize           uint32
	Next                uint32
	ExecutableEntry     uint64
	LoadAddress         uint64
	DataOffset          uint32
	Attributes          uint32
	NumSections         uint32
	ChecksumTableOffset uint32
	ImageHeaderOffset   uint32
	ACOffset            uint32
	ID                  uint32
	Checksum            uint32
}

type Partition struct {
	PartitionHeader
	Offset int64
	r      io.ReaderAt
}

type File struct {
	Header
	r          io.ReaderAt
	Registers  [256][2]uint32
	PUF        []byte
	ImageTable ImageHeaderTable
	Images     []*Image
	Partitions []*Partition
}

var sig = [4]byte{'X', 'N', 'L', 'X'}

func Open(name string) (*File, error) {
	r, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	f, err := NewFile(r)
	if err != nil {
		r.Close()
	}
	return f, err
}

func NewFile(r io.ReaderAt) (*File, error) {
	sr := io.NewSectionReader(r, 0, math.MaxUint32)

	f := &File{}
	h := &f.Header
	err := binary.Read(sr, binary.LittleEndian, h)
	if err != nil {
		return nil, err
	}

	if h.Sig != sig {
		return nil, fmt.Errorf("invalid boot signature: %q", h.Sig)
	}

	err = binary.Read(sr, binary.LittleEndian, &f.Registers)
	if err != nil {
		return nil, fmt.Errorf("failed to read register initialization table: %w", err)
	}

	if (h.FSBLImageFlags>>6)&3 == 3 {
		f.PUF = make([]byte, 1544)
		err = binary.Read(sr, binary.LittleEndian, &f.PUF)
		if err != nil {
			return nil, fmt.Errorf("failed to read PUF helper data: %w", err)
		}
	}

	ir := io.NewSectionReader(r, int64(h.ImageOffset), math.MaxUint32)
	err = binary.Read(ir, binary.LittleEndian, &f.ImageTable)
	if err != nil {
		return nil, fmt.Errorf("failed to read image header table: %w", err)
	}

	off := 4 * f.ImageTable.ImageOffset
	for {
		ir = io.NewSectionReader(r, int64(off), math.MaxUint32)

		var ih ImageHeader
		err = binary.Read(ir, binary.LittleEndian, &ih)
		if err != nil {
			return nil, fmt.Errorf("failed to read image header: %w", err)
		}

		f.Images = append(f.Images, &Image{ImageHeader: ih, Name: readstrz(ir)})

		off = 4 * ih.Next
		if off == 0 {
			break
		}
	}

	off = h.PartitionOffset
	for {
		pr := io.NewSectionReader(r, int64(off), math.MaxUint32)

		var ph PartitionHeader
		err = binary.Read(pr, binary.LittleEndian, &ph)
		if err != nil {
			return nil, fmt.Errorf("failed to read partition header: %w", err)
		}

		f.Partitions = append(f.Partitions, &Partition{PartitionHeader: ph, Offset: int64(off / 4), r: r})

		off = 4 * ph.Next
		if off == 0 {
			break
		}
	}

	return f, nil
}

func (f *File) Close() error {
	c, ok := f.r.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}

func (p *Partition) Open() io.Reader {
	return io.NewSectionReader(p.r, int64(4*p.DataOffset), int64(p.TotalSize))
}

func readstrz(r io.Reader) string {
	p := ""
loop:
	for {
		var b [4]byte
		err := binary.Read(r, binary.LittleEndian, &b)
		if err != nil {
			break
		}

		for i := 3; i >= 0; i-- {
			if b[i] == 0 {
				break loop
			}
			p += string(b[i])
		}

	}
	return p
}

type PartitionAttribute uint32

func (p PartitionAttribute) String() string {
	tr := "Non-secure"
	if p&1 != 0 {
		tr = "Secure"
	}

	el := (p >> 1) & 3
	ae := 64
	if (p>>3)&1 != 0 {
		ae = 32
	}

	enc := "Not Encrypted"
	if (p>>7)&1 != 0 {
		enc = "Encrypted"
	}

	cs := "No Checksum"
	switch (p >> 12) & 7 {
	case 1, 2, 4, 5, 6:
		cs = "Reserved Checksum"
	case 3:
		cs = "SHA3"
	}

	rs := "No Authentication"
	if (p>>15)&1 != 0 {
		rs = "RSA Authentication"
	}

	dd := "Unknown"
	switch (p >> 4) & 3 {
	case 0:
		dd = "None"
	case 1:
		dd = "PS"
	case 2:
		dd = "PL"
	}

	return fmt.Sprintf("Trustzone [%s], EL%d, AARCH%d, Device [%s], %s, %s, %s",
		tr, el, ae, dd, enc, cs, rs)
}

type KeySource uint32

func (p KeySource) String() string {
	switch p {
	case 0x00000000:
		return "Unencrypted Key Source"
	case 0xA5C3C5A5:
		return "Black Key stored in EFUSE"
	case 0xA5C3C5A7:
		return "Obfuscated key stored in eFUSE"
	case 0x3A5C3C5A:
		return "Red key stored in BBRAM"
	case 0xA5C3C5A3:
		return "eFUSE RED key stored in eFUSE"
	case 0xA35C7CA5:
		return "Obfuscated key stored in Boot Header"
	case 0xA3A5C3C5:
		return "USER key stored in Boot Header"
	case 0xA35C7C53:
		return "Black key stored in Boot Header"
	}
	return "Unknown Key Source"
}

type FSBLImageAttribute uint32

func (p FSBLImageAttribute) String() string {
	pf := "PUF in eFUSE"
	if (p>>6)&3 == 3 {
		pf = "PUF in boot header"
	}

	hs := "No Integrity Check"
	if (p>>8)&3 == 3 {
		hs = "SHA3 Integrity Check"
	}

	cs := ""
	switch (p >> 10) & 3 {
	case 0:
		cs = "R5 Single"
	case 1:
		cs = "A53 Single 32"
	case 2:
		cs = "A53 Single 64"
	case 3:
		cs = "R5 Dual"
	}

	rs := "RSA Authentication in eFUSE"
	if (p>>14)&3 == 3 {
		rs = "RSA Authentication in boot header"
	}

	return fmt.Sprintf("%s, %s, %s, %s", pf, hs, cs, rs)
}
