package fdt

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type Header struct {
	Magic           uint32
	Size            uint32
	StructOff       uint32
	StringsOff      uint32
	ReserveOff      uint32
	Version         uint32
	LastCompVersion uint32
	BootCpuid       uint32
	StringsSize     uint32
	StructSize      uint32
}

type Node struct {
	Prop []Prop
	Next *Node
}

type Prop struct {
	Name  string
	Value string
}

type Reserve struct {
	Addr uint64
	Size uint64
}

type File struct {
	Header
	Reserves []Reserve
	Nodes    []*Node
}

const (
	BEGIN_NODE = 0x1
	END_NODE   = 0x2
	PROP       = 0x3
	NOP        = 0x4
	END        = 0x9
)

var (
	ErrHeader = errors.New("fdt: invalid header")
)

func Decode(r io.ReaderAt) (*File, error) {
	sr := io.NewSectionReader(r, 0, math.MaxUint32)

	var hdr Header
	err := binary.Read(sr, binary.BigEndian, &hdr)
	if err != nil {
		return nil, wrapError(err)
	}

	if hdr.Magic != 0xd00dfeed {
		return nil, ErrHeader
	}

	var (
		resrvs []Reserve
		resrv  Reserve
	)
	sr.Seek(int64(hdr.ReserveOff), io.SeekStart)
	for {
		err := binary.Read(sr, binary.BigEndian, &resrv)
		if err != nil {
			return nil, wrapError(err)
		}
		if resrv.Addr == 0 && resrv.Size == 0 {
			break
		}
		resrvs = append(resrvs, resrv)
	}

	sr.Seek(int64(hdr.StructOff), io.SeekStart)
	lr := &io.LimitedReader{sr, int64(hdr.StructSize)}
	br := bufio.NewReader(lr)

loop:
	for {
		var typ uint32
		err := binary.Read(br, binary.BigEndian, &typ)
		if err != nil {
			return nil, wrapError(err)
		}

		switch typ {
		case BEGIN_NODE:
		case END_NODE:
		case PROP:
		case NOP:
		case END:
			break loop
		}
	}

	return &File{
		Header:   hdr,
		Reserves: resrvs,
	}, nil
}

func WriteDTS(w io.Writer, f *File) error {
	b := bufio.NewWriter(w)
	fmt.Fprintf(w, "/dts-v1/;\n")
	return wrapError(b.Flush())
}

func wrapError(err error) error {
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return fmt.Errorf("fdt: %v", err)
	}
	return nil
}
