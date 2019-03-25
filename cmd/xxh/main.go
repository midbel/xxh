package main

import (
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"time"

	"github.com/midbel/xxh"
)

func main() {
	kind := flag.Uint("k", 0, "hash")
	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()
	computeDigests(flag.Args(), *kind, *seed)
}

func computeDigests(files []string, kind, seed uint) {
	var (
		digest  hash.Hash
		pattern string
	)
	switch kind {
	case 0, 64:
		digest, pattern = xxh.New64(uint64(seed)), "%016x  %s - %dKB (%s)\n"
	case 32:
		digest, pattern = xxh.New32(uint32(seed)), "%08x %s - %dKB (%s)\n"
	default:
		return
	}

	buffer := make([]byte, 32<<10)
	for i := 0; i < len(files); i++ {
		r, err := os.Open(files[i])
		if err != nil {
			continue
		}

		when := time.Now()
		if n, err := io.CopyBuffer(digest, r, buffer); err == nil {
			sum := digest.Sum(nil)
			fmt.Printf(pattern, sum, files[i], n>>10, time.Since(when))
			digest.Reset()
		}
		r.Close()
	}
}
