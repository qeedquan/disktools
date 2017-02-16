package uimage

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	COMP_NONE  = iota // No	  Compression Used
	COMP_GZIP         // gzip  Compression Used
	COMP_BZIP2        // bzip2 Compression Used
	COMP_LZMA         // lzma  Compression Used
	COMP_LZO          // lzo   Compression Used
	COMP_LZ4          // lz4   Compression Used
	COMP_COUNT
)

const (
	TYPE_INVALID    = iota // Invalid Image
	TYPE_STANDALONE        // Standalone Program
	TYPE_KERNEL            // OS Kernel Image
	TYPE_RAMDISK           // RAMDisk Image
	TYPE_MULTI             // Multi-File Image
	TYPE_FIRMWARE          // Firmware Image
	TYPE_SCRIPT            // Script file
	TYPE_FILESYSTEM        // Filesystem Image (any type)
	TYPE_FLATDT            // Binary Flat Device Tree Blob
	TYPE_KWBIMAGE          // Kirkwood Boot Image
	TYPE_IMXIMAGE          // Freescale IMXBoot Image
	TYPE_UBLIMAGE          // Davinci UBL Image
	TYPE_OMAPIMAGE         // TI OMAP Config Header Image
	TYPE_AISIMAGE          // TI Davinci AIS Image
	// OS Kernel Image can run from any load address
	TYPE_KERNEL_NOLOAD
	TYPE_PBLIMAGE     // Freescale PBL Boot Image
	TYPE_MXSIMAGE     // Freescale MXSBoot Image
	TYPE_GPIMAGE      // TI Keystone GPHeader Image
	TYPE_ATMELIMAGE   // ATMEL ROM bootable Image
	TYPE_SOCFPGAIMAGE // Altera SOCFPGA Preloader
	TYPE_X86_SETUP    // x86 setup.bin Image
	TYPE_LPC32XXIMAGE // x86 setup.bin Image
	TYPE_LOADABLE     // A list of typeless images
	TYPE_RKIMAGE      // Rockchip Boot Image
	TYPE_RKSD         // Rockchip SD card
	TYPE_RKSPI        // Rockchip SPI image
	TYPE_ZYNQIMAGE    // Xilinx Zynq Boot Image
	TYPE_ZYNQMPIMAGE  // Xilinx ZynqMP Boot Image
	TYPE_FPGA         // FPGA Image
	TYPE_VYBRIDIMAGE  // VYBRID .vyb Image
	TYPE_TEE          // Trusted Execution Environment OS Image
	TYPE_FIRMWARE_IVT // Firmware Image with HABv4 IVT
	TYPE_COUNT        // Number of image types
)

const (
	ARCH_INVALID    = iota // Invalid CPU
	ARCH_ALPHA             // Alpha
	ARCH_ARM               // ARM
	ARCH_I386              // Intel x86
	ARCH_IA64              // IA64
	ARCH_MIPS              // MIPS
	ARCH_MIPS64            // MIPS	 64 Bit
	ARCH_PPC               // PowerPC
	ARCH_S390              // IBM S390
	ARCH_SH                // SuperH
	ARCH_SPARC             // Sparc
	ARCH_SPARC64           // Sparc 64 Bit
	ARCH_M68K              // M68K
	ARCH_NIOS              // Nios-32
	ARCH_MICROBLAZE        // MicroBlaze
	ARCH_NIOS2             // Nios-II
	ARCH_BLACKFIN          // Blackfin
	ARCH_AVR32             // AVR32
	ARCH_ST200             // STMicroelectronics ST200
	ARCH_SANDBOX           // Sandbox architecture (test only)
	ARCH_NDS32             // ANDES Technology - NDS32
	ARCH_OPENRISC          // OpenRISC 1000
	ARCH_ARM64             // ARM64
	ARCH_ARC               // Synopsys DesignWare ARC
	ARCH_X86_64            // AMD x86_64 Intel and Via
	ARCH_XTENSA            // Xtensas
	ARCH_COUNT
)

const (
	OS_INVALID   = iota // Invalid OS
	OS_OPENBSD          // OpenBSD
	OS_NETBSD           // NetBSD
	OS_FREEBSD          // FreeBSD
	OS_4_4BSD           // 4.4BSD
	OS_LINUX            // Linux
	OS_SVR4             // SVR4
	OS_ESIX             // Esix
	OS_SOLARIS          // Solaris
	OS_IRIX             // Irix
	OS_SCO              // SCO
	OS_DELL             // Dell
	OS_NCR              // NCR
	OS_LYNXOS           // LynxOS
	OS_VXWORKS          // VxWorks
	OS_PSOS             // pSOS
	OS_QNX              // QNX
	OS_U_BOOT           // Firmware
	OS_RTEMS            // RTEMS
	OS_ARTOS            // ARTOS
	OS_UNITY            // Unity OS
	OS_INTEGRITY        // INTEGRITY
	OS_OSE              // OSE
	OS_PLAN9            // Plan 9
	OS_OPENRTOS         // OpenRTOS
	OS_COUNT
)

type Header struct {
	Magic  uint32
	CRC    uint32
	Time   uint32
	Filesz uint32
	Load   uint32
	Entry  uint32
	DCRC   uint32
	OS     uint8
	Arch   uint8
	Type   uint8
	Comp   uint8
	Name   [32]byte
}

type File struct {
	*io.SectionReader
	Header
	Off int64
}

type Reader struct {
	File []*File
}

var (
	ErrHeader = errors.New("uimage: invalid header")
)

const magic = 0x27051956

func Open(r io.ReaderAt) ([]*File, error) {
	sr := io.NewSectionReader(r, 0, math.MaxInt32)

	var h Header
	err := binary.Read(sr, binary.BigEndian, &h)
	if err != nil {
		return nil, wrapError(err)
	}

	if h.Magic != magic {
		return nil, ErrHeader
	}

	var files []*File
	files = append(files, &File{
		Header: h,
		Off:    64,
	})

	if h.Type == TYPE_MULTI {
		var length uint32
		var off int64
		for {
			err = binary.Read(sr, binary.BigEndian, &length)
			if err != nil {
				return nil, wrapError(err)
			}
			if length == 0 {
				break
			}

			files = append(files, &File{
				Header: Header{Filesz: length},
				Off:    off,
			})
			off += int64(length)
		}
		files[0].Off = 0
		for _, f := range files {
			f.Off += 64 + 4*int64(len(files))
		}
	}

	for _, f := range files {
		f.SectionReader = io.NewSectionReader(r, f.Off, int64(f.Filesz))
	}

	return files, nil
}

func Write(w io.Writer, files []*File) error {
	if len(files) == 0 {
		return nil
	}

	b := bufio.NewWriter(w)

	binary.Write(b, binary.BigEndian, &files[0].Header)
	if len(files) > 0 {
		for _, f := range files[1:] {
			binary.Write(b, binary.BigEndian, f.Filesz)
		}
		binary.Write(b, binary.BigEndian, uint32(0))
	}

	for _, f := range files {
		f.Seek(0, io.SeekStart)
		io.Copy(b, f)
	}

	return wrapError(b.Flush())
}

func wrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("uimage: %v", err)
}
