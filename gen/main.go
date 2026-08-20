package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/go-archive"
	"github.com/moby/go-archive/compression"
)

func main() {
	name := filepath.Base(os.Args[0])

	var isImporting bool
	switch name {
	case "import":
		isImporting = true
	case "export":
		isImporting = false
	default:
		log.Fatalf("unknown mode: %s", name)
	}

	if isImporting {
		log.Println("Importing...")
	} else {
		log.Println("Exporting...")
	}

	for _, pair := range os.Args[1:] {
		arg1, arg2, ok := strings.Cut(pair, ":")
		if !ok {
			log.Fatalf("invalid pair: %q", pair)
		}

		log.Println(arg1, "->", arg2)
		var err error
		if isImporting {
			err = Import(arg1, arg2)
		} else {
			err = Export(arg1, arg2)
		}
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}

func Import(from, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	return archive.Untar(f, to, &archive.TarOptions{})
}

func Export(from, to string) error {
	r, err := archive.Tar(from, compression.Gzip)
	if err != nil {
		return err
	}
	defer r.Close()

	out, err := os.Create(to)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		os.Remove(to)
		return err
	}

	return out.Close()
}
