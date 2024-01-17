package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/qeedquan/disktools/tpm/v2/tpmtool"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	buf, err := os.ReadFile(flag.Arg(0))
	check(err)

	file, err := tpmtool.Decode(bytes.NewReader(buf))
	check(err)

	fmt.Printf("File Length %d\n", len(buf))
	dump(file)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: <file>")
	flag.PrintDefaults()
	os.Exit(2)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func dump(f *tpmtool.File) {
	fmt.Printf("Magic         %#x\n", f.Magic)
	fmt.Printf("Version       %d\n", f.Version)
	fmt.Printf("Hierarchy     %#x\n", f.Hierarchy)
	fmt.Printf("Saved Handle  %#x\n", f.SavedHandle)
	fmt.Printf("Sequence      %#x\n", f.Sequence)
	fmt.Printf("Length        %d\n", f.Length)
}
