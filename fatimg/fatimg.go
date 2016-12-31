package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/qeedquan/disktools/fat"
	"github.com/qeedquan/disktools/iod"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("fatimg: ")

	case_ := flag.Bool("c", false, "case sensitive")
	offset := flag.Int64("o", 0x7e00, "offset to start reading")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	fd, err := os.Open(flag.Arg(0))
	ck(err)

	rw := iod.NewORW(fd, *offset)
	opt := &fat.FileSystemOptions{
		Case: *case_,
	}
	_, err = fat.NewFileSystem(rw, opt)
	ck(err)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: fatimg [options] file")
	flag.PrintDefaults()
	os.Exit(2)
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
