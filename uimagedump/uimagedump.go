package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qeedquan/disktools/uimage"
)

var (
	outdir = flag.String("o", "", "output directory")

	status = 0
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	for _, name := range flag.Args() {
		ek(dump(name))
	}

	os.Exit(status)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: imagedump [options] file ...")
	flag.PrintDefaults()
	os.Exit(2)
}

func ek(err error) bool {
	if err != nil {
		fmt.Fprintln(os.Stderr, "uimagedump:", err)
		status = 1
		return true
	}
	return false
}

func dump(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := uimage.Open(f)
	if err != nil {
		return err
	}

	for _, p := range files {
		name := filepath.Join(*outdir, string(p.Name[:]))
		name = strings.TrimRight(name, "\x00")
		w, err := os.Create(name)
		if ek(err) {
			continue
		}
		_, err = io.Copy(w, p)
		ek(err)
		ek(w.Close())
	}

	return nil
}
