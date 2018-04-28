package hob

import "github.com/qeedquan/disktools/efi"

const (
	TYPE_END_OF_HOB_LIST = 0xffff
)

type Header struct {
	Type   uint16
	Length uint16
	_      uint32
}

type MemoryAllocationHeader struct {
	Name              efi.GUID
	MemoryBaseAddress efi.PhysAddr
	MemoryLength      uint64
	MemoryType        efi.MemoryType
	_                 [4]uint8
}

// PHIT
type HandoffInfoTable struct {
	Header              Header
	Version             uint32
	BootMode            efi.BootMode
	EfiMemoryTop        efi.PhysAddr
	EfiMemoryBottom     efi.PhysAddr
	EfiFreeMemoryTop    efi.PhysAddr
	EfiFreeMemoryBottom efi.PhysAddr
	EfiEndOfHobList     efi.PhysAddr
}

type MemoryAllocationModule struct {
	Header                 Header
	MemoryAllocationHeader MemoryAllocationHeader
	ModuleName             efi.GUID
	EntryPoint             efi.PhysAddr
}

type MemoryAllocationStack struct {
	Header          Header
	AllocDescriptor MemoryAllocationHeader
}

type MemoryAllocationBspStore struct {
	Header          Header
	AllocDescriptor MemoryAllocationHeader
}

type ResourceDescriptor struct {
	Header            Header
	Owner             efi.GUID
	ResourceType      efi.ResourceType
	ResourceAttribute efi.ResourceAttributeType
	PhysicalStart     efi.PhysAddr
	ResourceLength    uint64
}

type Cpu struct {
	Header            Header
	SizeOfMemorySpace uint8
	SizeOfIoSpace     uint8
	_                 [6]uint8
}

type CapsuleVolume struct {
	Header      Header
	BaseAddress efi.PhysAddr
	Length      uint64
}

type FirmwareVolume struct {
	Header      Header
	BaseAddress efi.PhysAddr
	Length      uint64
}
