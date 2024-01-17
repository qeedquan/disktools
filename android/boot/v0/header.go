package boot

const (
	MAGIC           = "ANDROID!"
	MAGIC_SIZE      = 8
	NAME_SIZE       = 16
	ARGS_SIZE       = 512
	EXTRA_ARGS_SIZE = 1024
)

type Header struct {
	// Must be BOOT_MAGIC.
	Magic         [MAGIC_SIZE]byte
	KernelSize    uint32 // size in bytes
	KernelAddr    uint32 // physical load addr
	RamdiskSize   uint32 // size in bytes
	RamdiskAddr   uint32 // physical load addr
	SecondSize    uint32 // size in bytes
	SecondAddr    uint32 // physical load addr
	TagsAddr      uint32 // physical addr for kernel tags (if required)
	PageSize      uint32 // flash page size we assume
	HeaderVersion uint32 // header version

	// Operating system version and security patch level.
	// For version "A.B.C" and patch level "Y-M-D":
	//   (7 bits for each of A, B, C; 7 bits for (Y-2000), 4 bits for M)
	//   os_version = A[31:25] B[24:18] C[17:11] (Y-2000)[10:4] M[3:0]
	OSVersion uint32

	Name    [NAME_SIZE]byte // asciiz product name
	Cmdline [ARGS_SIZE]byte // asciiz kernel commandline

	ID [8]uint32 // timestamp / checksum / sha1 / etc

	// Supplemental command line data; kept here to maintain
	// binary compatibility with older versions of mkbootimg.
	// Asciiz.
	ExtraCmdline [EXTRA_ARGS_SIZE]byte
}
