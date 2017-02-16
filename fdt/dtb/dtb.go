package dtb

type Header struct {
	Magic           uint32
	Size            uint32
	StructOff       uint32
	StringsOff      uint32
	MemOff          uint32
	Version         uint32
	LastCompVersion uint32
	BootCpuid       uint32
	StringsSize     uint32
	StructSize      uint32
}

type Reserve struct {
	Addr uint64
	Size uint64
}

const (
	BEGIN_NODE = 0x1
	END_NODE   = 0x2
	PROP       = 0x3
	NOP        = 0x4
	END        = 0x9
)
