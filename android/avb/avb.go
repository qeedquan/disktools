package avb

type Header struct {
	Magic                   [4]uint8
	Major                   uint32
	Minor                   uint32
	AuthBlockSize           uint32
	AuxBlockSize            uint32
	AlgorithmType           uint32
	HashOffset              uint64
	HashSize                uint64
	SignatureOffset         uint64
	PublicKeySize           uint64
	PublicKeyMetadataOffset uint64
	PublicKeyMetadataSize   uint64
	DescriptorsOffset       uint64
	DescriptorsSize         uint64
	RollbackIndex           uint64
	Flags                   uint32
	RollbackIndexLocation   uint32
	ReleaseString           [48]byte
	Reserved                [80]byte
}
