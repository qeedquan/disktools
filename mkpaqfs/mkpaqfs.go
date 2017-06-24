package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/qeedquan/disktools/paq"
)

func main() {
	var (
		outfile string
		options paq.WriteOptions
	)

	log.SetFlags(0)
	log.SetPrefix("mkpaqfs: ")

	flag.StringVar(&outfile, "o", "", "output file")
	flag.StringVar(&options.Label, "l", "", "label")
	flag.IntVar(&options.BlockSize, "b", 4096, "block size")
	flag.BoolVar(&options.Compress, "u", false, "no compression")
	flag.Usage = usage
	flag.Parse()

	out := bufio.NewWriter(os.Stdin)
	if outfile != "" {
		w, err := os.Create(outfile)
		ck(err)
		out = bufio.NewWriter(w)
		defer func() {
			ck(w.Close())
		}()
	}

	source := "."
	if flag.NArg() >= 1 {
		source = flag.Arg(0)
	}

	if options.Label == "" {
		options.Label = filepath.Base(source)
	}

	ck(mkpaqfs(out, source, &options))
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mkpaqfs [options] [source]")
	flag.PrintDefaults()
	os.Exit(2)
}

func mkpaqfs(out *bufio.Writer, source string, options *paq.WriteOptions) error {
	w, err := paq.NewWriter(out, options)
	if err != nil {
		return err
	}

	fd, err := os.Open(source)
	if err != nil {
		return err
	}

	fi, err := fd.Stat()
	if err != nil {
		return err
	}

	var pd *paq.Dir

	w.WriteHeader()
	if fi.IsDir() {
		pd, err = w.WriteDir(source, fi)
	} else {
		pd, err = w.WriteFile(source, fd)
	}
	if err != nil {
		return err
	}
	off := w.WriteBlockDir(pd)
	w.WriteTrailer(off)
	return w.Close()
}