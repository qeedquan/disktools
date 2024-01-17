package efi

type GUID [16]uint8
type Handle interface{}
type Status int
type Attributes2 uint32
type BootMode uint32
type PhysAddr uint64
type VirtAddr uint64
type ResourceType uint32
type ResourceAttributeType uint32
type MemoryType uint32

type TableHeader struct {
	Signature  uint64
	Revision   uint32
	HeaderSize uint32
	CRC32      uint32
	_          uint32
}

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

type Time struct {
	Year       uint16
	Month      uint8
	Day        uint8
	Hour       uint8
	Minute     uint8
	Second     uint8
	Pad1       uint8
	Nanosecond uint32
	Timezone   int16
	Daylight   uint8
	Pad2       uint8
}

type RawSection CommonSectionHeader
type RawSection2 CommonSectionHeader2

const (
	RESOURCE_SYSTEM_MEMORY         ResourceType = 0x00000000
	RESOURCE_MEMORY_MAPPED_IO      ResourceType = 0x00000001
	RESOURCE_IO                    ResourceType = 0x00000002
	RESOURCE_FIRMWARE_DEVICE       ResourceType = 0x00000003
	RESOURCE_MEMORY_MAPPED_IO_PORT ResourceType = 0x00000004
	RESOURCE_MEMORY_RESERVED       ResourceType = 0x00000005
	RESOURCE_IO_RESERVED           ResourceType = 0x00000006
	RESOURCE_MAX_MEMORY_TYPE       ResourceType = 0x00000007
)

const (
	BOOT_WITH_FULL_CONFIGURATION                  BootMode = 0x00
	BOOT_WITH_MINIMAL_CONFIGURATION               BootMode = 0x01
	BOOT_ASSUMING_NO_CONFIGURATION_CHANGES        BootMode = 0x02
	BOOT_WITH_FULL_CONFIGURATION_PLUS_DIAGNOSTICS BootMode = 0x03
	BOOT_WITH_DEFAULT_SETTINGS                    BootMode = 0x04
	BOOT_ON_S4_RESUME                             BootMode = 0x05
	BOOT_ON_S5_RESUME                             BootMode = 0x06
	BOOT_ON_S2_RESUME                             BootMode = 0x10
	BOOT_ON_S3_RESUME                             BootMode = 0x11
	BOOT_ON_FLASH_UPDATE                          BootMode = 0x12
	BOOT_IN_RECOVERY_MODE                         BootMode = 0x20
)

const (
	SUCCESS Status = 0
)

const (
	ReservedMemoryType      MemoryType = iota // Not used.
	LoaderCode                                // The code portions of a loaded application. (Note that UEFI OS loaders are UEFI applications.)
	LoaderData                                // The data portions of a loaded application and the default data allocation type used by an application to allocate pool memory.
	BootServicesCode                          // The code portions of a loaded Boot Services Driver.
	BootServicesData                          // The data portions of a loaded Boot Serves Driver, and the default data allocation type used by a Boot Services Driver to allocate pool memory.
	RuntimeServicesCode                       // The code portions of a loaded Runtime Services Driver.
	RuntimeServicesData                       // The data portions of a loaded Runtime Services Driver and the default data allocation type used by a Runtime Services Driver to allocate pool memory.
	ConventionalMemory                        // Free (unallocated) memory.
	UnusableMemory                            // Memory in which errors have been detected.
	ACPIReclaimMemory                         // Memory that holds the ACPI tables.
	ACPIMemoryNVS                             // Address space reserved for use by the firmware.
	MemoryMappedIO                            // Used by system firmware to request that a memory-mapped IO region be mapped by the OS to a virtual address so it can be accessed by EFI runtime services.
	MemoryMappedIOPortSpace                   // System memory-mapped IO region that is used to translate memory cycles to IO cycles by the processor. Note: There is only one region of type EfiMemoryMappedIoPortSpace defined in the architecture for Itanium-based platforms. As a result, there should be one and only one region of type EfiMemoryMappedIoPortSpace in the EFI memory map of an Itanium-based platform.
	PalCode                                   // Address space reserved by the firmware for code that is part of the processor.
	MaxMemoryType
)
