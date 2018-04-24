package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/qeedquan/disktools/efi"
	fsp "github.com/qeedquan/disktools/fsp/v1"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("fspdump: ")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	fp, err := readFSP(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	spew.Dump(fp.VolumnHeader)
	spew.Dump(fp.VolumnExtHeader)
	spew.Dump(fp.FfsHeader)
	spew.Dump(fp.FspHeader)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fspdump [options] file")
	flag.PrintDefaults()
	os.Exit(2)
}

func readFSP(name string) (*fsp.File, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var (
		vh  efi.VolumnHeader
		veh efi.VolumnExtHeader
		ffs efi.FfsFileHeader
		hdr fsp.Header
		fw  []byte
	)
	err = readVolumnHeader(fd, &vh, &veh)
	if err != nil {
		return nil, err
	}

	err = readFfs(fd, &vh, &veh, &ffs)
	if err != nil {
		return nil, err
	}

	off := int64(vh.ExtHeaderOffset) + int64(veh.ExtHeaderSize)
	off = (off + 7) &^ 7
	off += int64(reflect.TypeOf(efi.FfsFileHeader{}).Size())
	off += int64(reflect.TypeOf(efi.RawSection{}).Size())
	r := io.NewSectionReader(fd, off, math.MaxInt32)
	binary.Read(r, binary.LittleEndian, &hdr)
	if hdr.Signature != [4]uint8{'F', 'S', 'P', 'H'} {
		return nil, fmt.Errorf("invalid FSP signature")
	}

	fw, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read firmware: %v", err)
	}

	return &fsp.File{
		VolumnHeader:    vh,
		VolumnExtHeader: veh,
		FfsHeader:       ffs,
		FspHeader:       hdr,
		Firmware:        fw,
	}, nil
}

func readVolumnHeader(fd *os.File, vh *efi.VolumnHeader, veh *efi.VolumnExtHeader) error {
	r := io.NewSectionReader(fd, 0, math.MaxInt32)
	binary.Read(r, binary.LittleEndian, &vh.ZeroVector)
	binary.Read(r, binary.LittleEndian, &vh.FileSystemGuid)
	binary.Read(r, binary.LittleEndian, &vh.FvLength)
	binary.Read(r, binary.LittleEndian, &vh.Signature)
	binary.Read(r, binary.LittleEndian, &vh.Attributes)
	binary.Read(r, binary.LittleEndian, &vh.HeaderLength)
	binary.Read(r, binary.LittleEndian, &vh.Checksum)
	binary.Read(r, binary.LittleEndian, &vh.ExtHeaderOffset)
	var reserved [1]byte
	binary.Read(r, binary.LittleEndian, &reserved)
	binary.Read(r, binary.LittleEndian, &vh.Revision)

	if vh.Signature != 0x4856465f {
		return fmt.Errorf("invalid EFI volumn header signature")
	}

	for {
		var bm efi.BlockMapEntry
		binary.Read(r, binary.LittleEndian, &bm.NumBlocks)
		binary.Read(r, binary.LittleEndian, &bm.Length)
		if bm.NumBlocks == 0 && bm.Length == 0 {
			break
		}
		vh.BlockMap = append(vh.BlockMap, bm)
	}

	if vh.ExtHeaderOffset != 0 {
		r = io.NewSectionReader(fd, int64(vh.ExtHeaderOffset), math.MaxInt32)
		binary.Read(r, binary.LittleEndian, veh)
	}

	return nil
}

func readFfs(fd *os.File, vh *efi.VolumnHeader, veh *efi.VolumnExtHeader, ffs *efi.FfsFileHeader) error {
	off := int64(vh.ExtHeaderOffset) + int64(veh.ExtHeaderSize)
	off = (off + 7) &^ 7
	r := io.NewSectionReader(fd, off, math.MaxInt32)
	binary.Read(r, binary.LittleEndian, ffs)
	if ffs.Name != [16]byte{0xbe, 0x40, 0x27, 0x91, 0x84, 0x22, 0x34, 0x47, 0xb9, 0x71, 0x84, 0xb0, 0x27, 0x35, 0x3f, 0x0c} {
		return fmt.Errorf("invalid FFS GUID")
	}
	return nil
}
