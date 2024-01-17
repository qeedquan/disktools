package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/qeedquan/disktools/xilinx/zynqmp/zynqmpboot"
)

var (
	outdir = flag.String("o", "", "output directory")
)

func main() {
	log.SetPrefix("xilinxbootdump: ")
	log.SetFlags(0)

	parseflags()
	xf, err := zynqmpboot.Open(flag.Arg(0))
	ck(err)
	defer xf.Close()

	info(xf)

	if *outdir == "" {
		return
	}

	os.MkdirAll(*outdir, 0755)
	for i, p := range xf.Partitions {
		r := p.Open()
		name := fmt.Sprintf("partition_%d", i)
		name = filepath.Join(*outdir, name)
		f, err := os.Create(name)
		if ek(err) {
			continue
		}
		_, err = io.Copy(f, r)
		ek(err)
		ek(f.Close())
	}
}

func ek(err error) bool {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true
	}
	return false
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseflags() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: [options] boot.bin")
	flag.PrintDefaults()
	os.Exit(2)
}

func info(xf *zynqmpboot.File) {
	h := &xf.Header
	fmt.Printf("Boot Header\n")
	fmt.Printf("-----------\n")
	for i, v := range h.Vectors {
		fmt.Printf("Vector #%d %#x\n", i, v)
	}
	fmt.Printf("Width Detection Word   %#x\n", h.Width)
	fmt.Printf("Key Source             %#x (%s)\n", h.KeySource, zynqmpboot.KeySource(h.KeySource))
	fmt.Printf("FSBL Execution Address %#x\n", h.FSBLEntry)
	fmt.Printf("Source Offset          %#x\n", h.SourceOffset)
	fmt.Printf("PMU Image Length       %d\n", h.PMUImageSize)
	fmt.Printf("PMU Total Length       %d\n", h.PMUTotalSize)
	fmt.Printf("FSBL Image Length      %d\n", h.FSBLImageSize)
	fmt.Printf("FSBL Total Length      %d\n", h.FSBLTotalSize)
	fmt.Printf("FSBL Image Flags       %#x (%s)\n", h.FSBLImageFlags, zynqmpboot.FSBLImageAttribute(h.FSBLImageFlags))
	fmt.Printf("Checksum               %#x\n", h.Checksum)
	fmt.Printf("Black Key              %#x\n", h.BlackKey)
	fmt.Printf("Shutter Value          %#x\n", h.ShutterValue)
	fmt.Printf("Image Offset           %#x\n", h.ImageOffset)
	fmt.Printf("Partition Offset       %#x\n", h.PartitionOffset)
	fmt.Printf("Secure IV              %x\n", h.SecureIV)
	fmt.Printf("Black Key IV           %x\n", h.BlackKeyIV)
	fmt.Printf("\n")

	if xf.PUF != nil {
		fmt.Printf("PUF Helper Data\n")
		fmt.Printf("---------------\n")
	} else {
		fmt.Printf("No PUF Helper data\n")
		fmt.Printf("------------------\n")
	}
	fmt.Printf("\n")

	fmt.Printf("Register Initialization Table\n")
	fmt.Printf("-----------------------------\n")
	for _, r := range xf.Registers {
		if r[0] != 0xFFFFFFFF {
			fmt.Printf("%#x %#x\n", r[0], r[1])
		}
	}
	fmt.Printf("\n")

	ih := &xf.ImageTable
	fmt.Printf("Image Header Table\n")
	fmt.Printf("------------\n")
	fmt.Printf("Version                 %#x\n", ih.Version)
	fmt.Printf("Number of Image Headers %d\n", ih.NumImages)
	fmt.Printf("Partition Offset        %#x\n", ih.PartitionOffset)
	fmt.Printf("Image Offset            %#x\n", ih.ImageOffset)
	fmt.Printf("Auth Offset             %#x\n", ih.AuthOffset)
	fmt.Printf("Secondary Boot Device   %#x\n", ih.SecondaryBootDevice)
	fmt.Printf("\n")

	fmt.Printf("Images Header\n")
	fmt.Printf("-------------\n")
	for i, p := range xf.Images {
		fmt.Printf("Image %d\n", i)
		fmt.Printf("Partition Offset     %#x\n", p.Partition)
		fmt.Printf("Number of Partitions %d\n", p.NumPartitions)
		fmt.Printf("Name                 %q\n", p.Name)
		fmt.Printf("\n")
	}
	fmt.Printf("\n")

	fmt.Printf("Partitions Header\n")
	fmt.Printf("----------------\n")
	for i, p := range xf.Partitions {
		fmt.Printf("Partition %d\n", i)
		fmt.Printf("Offset           %#x\n", p.Offset)
		fmt.Printf("Encrypted Size   %d\n", p.EncryptedSize)
		fmt.Printf("Unencrypted Size %d\n", p.UnencryptedSize)
		fmt.Printf("Total Size       %d\n", p.TotalSize)
		fmt.Printf("Executable Entry %#x\n", p.ExecutableEntry)
		fmt.Printf("Load Address     %#x\n", p.LoadAddress)
		fmt.Printf("Attributes       %#x (%s)\n", p.Attributes, zynqmpboot.PartitionAttribute(p.Attributes))
		fmt.Printf("Sections         %d\n", p.NumSections)
		fmt.Printf("ID               %#x\n", p.ID)
		fmt.Printf("Data Offset      %#x\n", p.DataOffset)
		fmt.Printf("AC Offset        %#x\n", p.ACOffset)
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}
