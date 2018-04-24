package efi

type GUID [16]uint8
type Attributes2 uint32

type BlockMapEntry struct {
	NumBlocks uint32
	Length    uint32
}

type VolumnHeader struct {
	ZeroVector      [16]uint8
	FileSystemGuid  GUID
	FvLength        uint64
	Signature       uint32
	Attributes      Attributes2
	HeaderLength    uint16
	Checksum        uint16
	ExtHeaderOffset uint16
	_               uint8
	Revision        uint8
	BlockMap        []BlockMapEntry
}

type VolumnExtHeader struct {
	FvName        GUID
	ExtHeaderSize uint32
}

type FfsFileHeader struct {
	Name           GUID
	IntegrityCheck uint16
	Type           uint8
	Attributes     uint8
	Size           [3]uint8
	State          uint8
}

type SectionType uint8

type CommonSectionHeader struct {
	Size [3]uint8
	Type SectionType
}

type CommonSectionHeader2 struct {
	Size         [3]uint8
	Type         SectionType
	ExtendedSize uint32
}

type RawSection CommonSectionHeader
type RawSection2 CommonSectionHeader2