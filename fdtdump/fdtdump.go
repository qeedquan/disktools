package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/qeedquan/disktools/devicetree/fdt"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("fdtdump: ")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	f, err := os.Open(flag.Arg(0))
	ck(err)
	defer f.Close()

	d, err := fdt.Decode(f)
	ck(err)

	ck(fdt.WriteDTS(os.Stdout, d))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: <dtb file>")
	flag.PrintDefaults()
	os.Exit(2)
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
