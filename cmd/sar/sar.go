package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/qeedquan/disktools/ar"
)

var (
	command   rune
	verbose   bool
	whitelist = make(map[string]bool)
)

func main() {
	log.SetPrefix("sar: ")
	log.SetFlags(0)
	parseFlags()
	runCommand(command)
}

func parseFlags() {
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
	}

	for _, ch := range flag.Arg(0) {
		switch ch {
		case 't', 'x':
			if command != 0 {
				log.Fatal("different operation options specified")
			}
			command = ch
		case 'v':
			verbose = true
		default:
			log.Fatalf("invalid option -- %q", ch)
		}
	}

	for i := 2; i < flag.NArg(); i++ {
		whitelist[flag.Arg(i)] = true
	}
}

func runCommand(command rune) {
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stdout, "sar: no error")
		return
	}

	fd, err := os.Open(flag.Arg(1))
	ck(err)
	defer fd.Close()

	r, err := ar.NewReader(fd)
	ck(err)

	switch command {
	case 't':
		list(r)
	case 'x':
		extract(r)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: sar [options] {tx}[v] archive-file file...")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, " commands:")
	fmt.Fprintln(os.Stderr, "  t        - display contents of archive")
	fmt.Fprintln(os.Stderr, "  x        - extract file(s) from the archive")
	fmt.Fprintln(os.Stderr, " generic modifiers:")
	fmt.Fprintln(os.Stderr, "  v        - be verbose")
	os.Exit(2)
}

func ek(err error) bool {
	if err != nil {
		fmt.Fprintln(os.Stderr, "sar:", err)
		return true
	}
	return false
}

func ck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func allowed(name string) bool {
	if name == "" {
		return false
	}
	if flag.NArg() < 3 {
		return true
	}
	return whitelist[name]
}

func list(r *ar.Reader) {
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		ck(err)

		if !allowed(h.Name) {
			continue
		}
		if verbose {
			size := fmt.Sprint(h.Size)
			fmt.Printf("%v %10s %v %s\n", h.Mode, size, h.Mtime, h.Name)
		} else {
			fmt.Println(h.Name)
		}
	}
}

func extract(r *ar.Reader) {
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		ck(err)

		if !allowed(h.Name) {
			continue
		}

		if verbose {
			fmt.Printf("x - %s\n", h.Name)
		}

		w, err := os.Create(h.Name)
		if ek(err) {
			continue
		}

		_, err = io.Copy(w, r)
		ek(err)
		ek(w.Close())
	}
}
