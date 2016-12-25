package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/qeedquan/disktools/endian"
	"github.com/qeedquan/disktools/gpt"
	"github.com/qeedquan/disktools/mbr"
)

func main() {
	log.SetFlags(0)

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	name := flag.Arg(0)
	fi, err := os.Stat(name)
	ck(err)
	if fi.IsDir() {
		ck(fmt.Errorf("%v: is a directory", name))
	}

	parts, err := discover(name)
	ck(err)

	if len(parts) == 0 {
		fmt.Println("No partitions found")
	}

	for _, p := range parts {
		switch p := p.(type) {
		case *mbr.Record:
			printMBR(p)
		case *gpt.Table:
			printGPT(p)
		default:
			panic("unreachable")
		}
		fmt.Println()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: disktool [options] file")
	flag.PrintDefaults()
	os.Exit(2)
}

func ck(err error) {
	if err != nil {
		log.Fatal("disktool: ", err)
	}
}

func discover(name string) ([]interface{}, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var parts []interface{}
	mbr, err := mbr.Open(f)
	if err == nil {
		parts = append(parts, mbr)
	}

	gpt, err := gpt.Open(f, nil)
	if err == nil {
		parts = append(parts, gpt)
	}

	return parts, nil
}

func printSpaces() {
	fmt.Printf("%s\n", strings.Repeat("-", 80))
}

func printMBR(p *mbr.Record) {
	empty := func(i int) bool {
		c := &p.Part[i]
		return c.FirstLBA == 0 || c.NumSectors == 0
	}

	protective := false
	if p.Part[0].Type == 0xee && !empty(0) && empty(1) && empty(2) && empty(3) {
		protective = true
	}

	if !protective {
		fmt.Printf("MBR Partition\n")
	} else {
		fmt.Printf("MBR Partition (Protective)\n")
	}

	printSpaces()
	if len(p.Part) == 0 {
		fmt.Println("No active partitions")
	}

	for i, c := range p.Part {
		size := (c.LastLBA - uint64(c.FirstLBA) + 1)

		fmt.Printf("Partition %d\n", i+1)
		fmt.Printf("  Bootable:       %b\n", c.Bootable)
		fmt.Printf("  Type:           %#x (%s)\n", c.Type, mbr.Types[c.Type])
		fmt.Printf("  First sector:   %d (%#x) (at %s)\n",
			c.FirstLBA, c.FirstLBA*uint32(p.Sectsz), endian.IEEE1541frombits(uint64(c.FirstLBA)*uint64(p.Sectsz)))
		fmt.Printf("  Last sector:    %d (%#x) (at %s)\n",
			c.LastLBA, c.LastLBA*uint64(p.Sectsz), endian.IEEE1541frombits(uint64(c.LastLBA)*uint64(p.Sectsz)))
		fmt.Printf("  Partition size: %d sectors (%s)\n",
			size, endian.IEEE1541frombits(size*uint64(p.Sectsz)))
		fmt.Printf("  Start CHS:      (%d,%d,%d)\n",
			c.Start.Head, c.Start.Sector, c.Start.Cylinder)
		fmt.Printf("  End CHS:        (%d,%d,%d)\n",
			c.End.Head, c.End.Sector, c.End.Cylinder)
		fmt.Printf("\n")
	}
	printSpaces()
}

func printGPT(p *gpt.Table) {
	h := &p.Header

	fmt.Printf("GPT Partition\n")
	printSpaces()
	fmt.Printf("Header\n")
	fmt.Printf("  GUID:        %v\n", h.GUID)
	fmt.Printf("  Header size: %v\n", h.Hdrsz)
	fmt.Printf("  Header CRC:  %#x\n", h.Hdrcrc)
	fmt.Printf("  Current LBA: %v\n", h.Current)
	fmt.Printf("  Backup LBA:  %v\n", h.Backup)
	fmt.Printf("  First LBA:   %v\n", h.First)
	fmt.Printf("  Last LBA:    %v\n", h.Last)
	fmt.Printf("  Entries:     %v\n", h.Ent)
	fmt.Printf("  Entrie size: %v\n", h.Entsz)
	fmt.Printf("  Table CRC:   %#x\n", h.Tabcrc)
	printSpaces()
}
