package main

import (
	"flag"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/midbel/xxh"
)

func main() {
	kind := flag.Uint("k", 0, "hash")
	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()

	var (
		digest  hash.Hash
		pattern string
	)
	switch *kind {
	case 0, 64:
		digest, pattern = xxh.New64(uint64(*seed)), "%016x %s\n"
	case 32:
		digest, pattern = xxh.New32(uint32(*seed)), "%08x %s\n"
	default:
		return
	}

	for _, f := range flag.Args() {
		r, err := os.Open(f)
		if err != nil {
			continue
		}

		if _, err := io.Copy(digest, r); err == nil {
			fmt.Printf(pattern, digest.Sum(nil), f)
			digest.Reset()
		}
		r.Close()
	}
}
