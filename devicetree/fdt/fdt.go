package fdt

import (
	"bufio"
	"bytes"
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
	Name     string
	Prop     []Prop
	Parent   *Node
	Children []*Node
}

type Prop struct {
	Name  string
	Value []byte
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
		str, nr, err := readString(br, false)
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
	var depth int
loop:
	for {
		var typ uint32
		err := binary.Read(br, binary.BigEndian, &typ)
		if err != nil {
			return nil, wrapError(err)
		}

		switch typ {
		case BEGIN_NODE:
			node := &Node{}
			node.Name, _, err = readString(br, true)
			if err != nil {
				return nil, wrapError(err)
			}
			if root == nil {
				root, cur = node, node
			} else {
				cur.Children = append(cur.Children, node)
				node.Parent = cur
				cur = node
			}
			depth++

		case END_NODE:
			if depth--; depth < 0 {
				return nil, fmt.Errorf("fdt: unbalanced begin_node/end_node token")
			}
			cur = cur.Parent

		case PROP:
			if depth <= 0 {
				return nil, fmt.Errorf("fdt: encountered prop outside of begin_node")
			}
			prop, err := readProp(br, stringtab)
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
	}
	if depth != 0 {
		return nil, fmt.Errorf("fdt: unbalanced begin_node/end_node token")
	}

	return &File{
		Header:   hdr,
		Reserves: resrvs,
		Root:     root,
		Strings:  stringtab,
	}, nil
}

func readString(b *bufio.Reader, padding bool) (string, int, error) {
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

	return p.String(), n, discardPad(padding, b, int64(n))
}

func readProp(b *bufio.Reader, tab map[int64]string) (prop Prop, err error) {
	var phdr struct {
		Len     uint32
		NameOff uint32
	}
	err = binary.Read(b, binary.BigEndian, &phdr)
	if err != nil {
		return
	}

	buf := make([]byte, phdr.Len)
	_, err = io.ReadAtLeast(b, buf, len(buf))
	if err != nil {
		return
	}

	prop = Prop{
		Name:  tab[int64(phdr.NameOff)],
		Value: buf,
	}

	err = discardPad(phdr.Len > 0, b, int64(phdr.Len))

	return
}

func discardPad(cond bool, b *bufio.Reader, n int64) error {
	if !cond {
		return nil
	}

	var discard [4]byte
	padsz := (4 - int(n)%4) % 4
	_, err := io.ReadAtLeast(b, discard[:padsz], padsz)
	if err != nil {
		return err
	}

	return nil
}

func WriteDTS(w io.Writer, f *File) error {
	b := bufio.NewWriter(w)
	fmt.Fprintf(b, "/dts-v1/;\n")
	fmt.Fprintf(b, "// magic:             %#x\n", f.Magic)
	fmt.Fprintf(b, "// totalsize:         %#x (%d)\n", f.Size, f.Size)
	fmt.Fprintf(b, "// off_dt_struct:     %#x\n", f.StructOff)
	fmt.Fprintf(b, "// off_dt_strings:    %#x\n", f.StringsOff)
	fmt.Fprintf(b, "// version:           %d\n", f.Version)
	fmt.Fprintf(b, "// last_comp_version: %d\n", f.LastCompVersion)
	if f.Version >= 2 {
		fmt.Fprintf(b, "// boot_cpuid_phys:   %#x\n", f.BootCpuid)
	}
	if f.Version >= 3 {
		fmt.Fprintf(b, "// size_dt_strings:   %#x\n", f.StringsSize)
	}
	if f.Version >= 17 {
		fmt.Fprintf(b, "// size_dt_struct:    %#x\n", f.StructSize)
	}

	for _, p := range f.Reserves {
		fmt.Fprintf(b, "/memreserve/ %x %x\n", p.Addr, p.Size)
	}

	fmt.Fprintf(b, "\n")
	writeStruct(b, f.Root, 0)

	return wrapError(b.Flush())
}

func writeStruct(b *bufio.Writer, node *Node, depth int) {
	const shift = 4

	if node == nil {
		return
	}
	s := node.Name
	if s == "" {
		s = "/"
	}
	fmt.Fprintf(b, "%*s%s {\n", depth*shift, "", s)

	depth++
	for _, p := range node.Prop {
		fmt.Fprintf(b, "%*s%s\n", depth*shift, "", p.Name)
	}
	for _, p := range node.Children {
		writeStruct(b, p, depth)
	}
	depth--
	fmt.Fprintf(b, "%*s}\n", depth*shift, "")
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
