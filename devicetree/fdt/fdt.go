package fdt

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
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
	Name string
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
	Root     *Node
	Strings  map[int64]string
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

	var stringtab = make(map[int64]string)
	sr.Seek(int64(hdr.StringsOff), io.SeekStart)
	lr := &io.LimitedReader{sr, int64(hdr.StringsSize)}
	br := bufio.NewReader(lr)
	off := int64(0)
	for {
		str, nr, err := readString(br)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, wrapError(err)
		}
		stringtab[off] = str
		off += int64(nr)
	}

	sr.Seek(int64(hdr.StructOff), io.SeekStart)
	lr = &io.LimitedReader{sr, int64(hdr.StructSize)}
	br = bufio.NewReader(lr)

	var root, cur *Node
	var link int
loop:
	for {
		var typ uint32
		err := binary.Read(br, binary.BigEndian, &typ)
		if err != nil {
			return nil, wrapError(err)
		}

		switch typ {
		case BEGIN_NODE:
			next := &Node{}
			next.Name, _, err = readString(br)
			if err != nil {
				return nil, wrapError(err)
			}
			if root == nil {
				root, cur = next, next
			} else {
				cur.Next, cur = next, next
			}
			link++

		case END_NODE:
			if link--; link < 0 {
				return nil, fmt.Errorf("fdt: unbalanced begin_node/end_node token")
			}

		case PROP:
			prop, err := readProp(br)
			if err != nil {
				return nil, wrapError(err)
			}
			cur.Prop = append(cur.Prop, prop)

		case NOP:

		case END:
			break loop

		default:
			return nil, fmt.Errorf("fdt: unknown node type %#x", typ)
		}

		_, err = skipPadding(br)
		if err != nil {
			return nil, wrapError(err)
		}
	}
	if link != 0 {
		return nil, fmt.Errorf("fdt: unbalanced begin_node/end_node token")
	}

	return &File{
		Header:   hdr,
		Reserves: resrvs,
		Root:     root,
		Strings:  stringtab,
	}, nil
}

func readString(b *bufio.Reader) (string, int, error) {
	p := new(bytes.Buffer)
	n := 0
	for {
		ch, sz, err := b.ReadRune()
		if err != nil {
			return p.String(), n, err
		}
		n += sz
		if ch == 0 {
			break
		}
		p.WriteRune(ch)
	}

	return p.String(), n, nil
}

func readProp(b *bufio.Reader) (prop Prop, err error) {
	var propsz [2]uint32
	err = binary.Read(b, binary.BigEndian, &propsz)
	if err != nil {
		return
	}

	if propsz[0] == 0 {
		return
	}

	buf := make([]byte, propsz[0])
	_, err = io.ReadAtLeast(b, buf, len(buf))
	if err != nil {
		return
	}

	prop = Prop{
		Value: strings.TrimRight(string(buf), "\x00"),
	}

	return
}

func skipPadding(b *bufio.Reader) (int, error) {
	n := 0
	for {
		buf, err := b.Peek(1)
		if err != nil {
			return n, err
		}
		if buf[0] != 0 {
			break
		}
		b.ReadByte()
		n++
	}
	return n, nil
}

func WriteDTS(w io.Writer, f *File) error {
	b := bufio.NewWriter(w)
	fmt.Fprintf(w, "/dts-v1/;\n")
	fmt.Fprintf(w, "// magic:             %#x\n", f.Magic)
	fmt.Fprintf(w, "// totalsize:         %#x (%d)\n", f.Size, f.Size)
	fmt.Fprintf(w, "// off_dt_struct:     %#x\n", f.StructOff)
	fmt.Fprintf(w, "// off_dt_strings:    %#x\n", f.StringsOff)
	fmt.Fprintf(w, "// version:           %d\n", f.Version)
	fmt.Fprintf(w, "// last_comp_version: %d\n", f.LastCompVersion)
	fmt.Fprintf(w, "// boot_cpuid_phys:   %#x\n", f.BootCpuid)
	fmt.Fprintf(w, "// size_dt_strings:   %#x\n", f.StringsSize)
	fmt.Fprintf(w, "// size_dt_struct:    %#x\n", f.StructSize)
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
