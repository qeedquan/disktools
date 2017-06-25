package paq

type Header struct {
	Magic     uint32
	Version   uint16
	BlockSize uint32
	Time      uint32
	Label     [32]byte
}

type Block struct {
	Magic    uint32
	Size     uint32
	Type     uint8
	Encoding uint8
	Adler32  uint32
}

type Trailer struct {
	Magic uint32
	Root  uint32
	Sha1  [20]byte
}

type Dir struct {
	Qid    uint32
	Mode   uint32
	Mtime  uint32
	Length uint32
	Offset uint32
	Name   string
	Uid    string
	Gid    string
}

const (
	dmdir = 0x80000000
)

const (
	DirBlock = iota
	DataBlock
	PointerBlock
)

const (
	NoEnc = iota
	DeflateEnc
)

const (
	HeaderMagic    = 0x529ab12b
	HeaderSize     = 44
	BigHeaderMagic = 0x25a9
	BlockMagic     = 0x198a1cbf
	BlockSize      = 12
	BigBlockMagic  = 0x91a8
	TrailerMagic   = 0x6b46e688
	TrailerSize    = 28
	Version        = 1
	MaxBlockSize   = 512 * 1024
	MinBlockSize   = 512
	MinDirSize     = 28
)
