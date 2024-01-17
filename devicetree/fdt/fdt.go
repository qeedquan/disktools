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
	"unicode"
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
	Value interface{}
}

type Reserve struct {
	Addr uint64
	Size uint64
}

type File struct {
	Header
	Reserves []Reserve
	Root     *Node
	Strings  []string
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

	var stringtab []string
	sr.Seek(int64(hdr.StringsOff), io.SeekStart)
	lr := &io.LimitedReader{sr, int64(hdr.StringsSize)}
	br := bufio.NewReader(lr)
	for {
		str, err := readString(br, false)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, wrapError(err)
		}
		stringtab = append(stringtab, str)
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
			node.Name, err = readString(br, true)
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
			prop, err := readProp(&hdr, r, br)
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

func readString(br *bufio.Reader, padding bool) (string, error) {
	p := new(bytes.Buffer)
	n := 0
	for {
		ch, sz, err := br.ReadRune()
		if err != nil {
			return p.String(), err
		}
		n += sz
		if ch == 0 {
			break
		}
		p.WriteRune(ch)
	}

	return p.String(), discardPad(padding, br, int64(n))
}

func readProp(hdr *Header, r io.ReaderAt, br *bufio.Reader) (prop Prop, err error) {
	var phdr struct {
		Len     uint32
		NameOff uint32
	}
	err = binary.Read(br, binary.BigEndian, &phdr)
	if err != nil {
		return
	}

	buf := make([]byte, phdr.Len)
	_, err = io.ReadAtLeast(br, buf, len(buf))
	if err != nil {
		return
	}

	sr := io.NewSectionReader(r, int64(hdr.StringsOff)+int64(phdr.NameOff), int64(hdr.StringsSize))
	name, err := readString(bufio.NewReader(sr), false)
	if err != nil {
		return
	}

	var value interface{}
	switch {
	case isPrint(buf):
		fields := strings.Split(string(buf), "\x00")
		for i := len(fields) - 1; i >= 0; i-- {
			if fields[i] != "" {
				fields = fields[:i+1]
				break
			}
		}

		if len(fields) > 1 {
			value = fields
		} else {
			value = fields[0]
		}

	case len(buf)%4 == 0:
		if len(buf) == 0 {
			value = nil
			break
		}

		var v []uint32
		for i := 0; i < len(buf); i += 4 {
			v = append(v, read4(buf[i:]))
		}
		value = v

	default:
		value = buf
	}

	prop = Prop{
		Name:  name,
		Value: value,
	}

	err = discardPad(phdr.Len > 0, br, int64(phdr.Len))

	return
}

func read4(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func isPrint(buf []byte) bool {
	if len(buf) == 0 {
		return false
	}
	if buf[len(buf)-1] != 0 {
		return false
	}

	seen := false
	count := 0
	r := bytes.NewReader(buf)
	for {
		ch, _, err := r.ReadRune()
		if err != nil {
			break
		}
		if ch == 0 && count != 0 {
			count = 0
			continue
		}
		if !unicode.IsPrint(ch) {
			return false
		}
		count++
		seen = true
	}
	return seen
}

func discardPad(cond bool, br *bufio.Reader, n int64) error {
	if !cond {
		return nil
	}

	var discard [4]byte
	padsz := (4 - int(n)%4) % 4
	_, err := io.ReadAtLeast(br, discard[:padsz], padsz)
	if err != nil {
		return err
	}

	return nil
}

func WriteDTS(w io.Writer, f *File) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "/dts-v1/;\n")
	fmt.Fprintf(bw, "// magic:             %#x\n", f.Magic)
	fmt.Fprintf(bw, "// totalsize:         %#x (%d)\n", f.Size, f.Size)
	fmt.Fprintf(bw, "// off_dt_struct:     %#x\n", f.StructOff)
	fmt.Fprintf(bw, "// off_dt_strings:    %#x\n", f.StringsOff)
	fmt.Fprintf(bw, "// version:           %d\n", f.Version)
	fmt.Fprintf(bw, "// last_comp_version: %d\n", f.LastCompVersion)
	if f.Version >= 2 {
		fmt.Fprintf(bw, "// boot_cpuid_phys:   %#x\n", f.BootCpuid)
	}
	if f.Version >= 3 {
		fmt.Fprintf(bw, "// size_dt_strings:   %#x\n", f.StringsSize)
	}
	if f.Version >= 17 {
		fmt.Fprintf(bw, "// size_dt_struct:    %#x\n", f.StructSize)
	}

	for _, p := range f.Reserves {
		fmt.Fprintf(bw, "/memreserve/ %x %x\n", p.Addr, p.Size)
	}

	fmt.Fprintf(bw, "\n")
	writeStruct(bw, f.Root, 0)

	return wrapError(bw.Flush())
}

func writeStruct(bw *bufio.Writer, node *Node, depth int) {
	const shift = 4

	if node == nil {
		return
	}
	s := node.Name
	if s == "" {
		s = "/"
	}
	fmt.Fprintf(bw, "%*s%s {\n", depth*shift, "", s)

	depth++
	for _, p := range node.Prop {
		fmt.Fprintf(bw, "%*s%s", depth*shift, "", p.Name)
		if p.Value != nil {
			fmt.Fprintf(bw, " = ")
		}

		switch v := p.Value.(type) {
		case []string:
			for i := range v {
				fmt.Fprintf(bw, "%q", v[i])
				if i+1 < len(v) {
					fmt.Fprintf(bw, ", ")
				}
			}

		case string:
			fmt.Fprintf(bw, "%q", v)

		case []byte:
			fmt.Fprintf(bw, "[")
			for i := range v {
				fmt.Fprintf(bw, "%02x", v[i])
				if i+1 < len(v) {
					fmt.Fprintf(bw, " ")
				}
			}
			fmt.Fprintf(bw, "]")

		case []uint32:
			fmt.Fprintf(bw, "<")
			for i := range v {
				fmt.Fprintf(bw, "%#08x", v[i])
				if i+1 < len(v) {
					fmt.Fprintf(bw, " ")
				}
			}
			fmt.Fprintf(bw, ">")

		case nil:

		default:
			fmt.Fprintf(bw, "(%T)", v)
		}
		fmt.Fprintf(bw, ";\n")
	}
	for _, p := range node.Children {
		writeStruct(bw, p, depth)
	}
	depth--
	fmt.Fprintf(bw, "%*s};\n", depth*shift, "")
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
