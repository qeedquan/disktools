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

	offset := flag.Int64("o", 0, "offset to start reading")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	fd, err := os.Open(flag.Arg(0))
	ck(err)

	rw := iod.NewORW(fd, *offset)
	_, err = fat.NewFileSystem(rw)
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
