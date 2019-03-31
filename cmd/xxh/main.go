package main

import (
	"flag"
	"fmt"
	"hash"
	"io"
	"os"
	"runtime/pprof"
	"time"

	"github.com/midbel/xxh"
)

func main() {
	cpu := flag.String("c", "", "cpu profile")
	kind := flag.Uint("k", 0, "hash")
	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()

	if *cpu != "" {
		w, err := os.Create(*cpu)
		if err != nil {
			os.Exit(5)
		}
		defer w.Close()
		if err := pprof.StartCPUProfile(w); err != nil {
			os.Exit(7)
		}
		defer pprof.StopCPUProfile()
	}

	computeDigests(flag.Args(), *kind, *seed)
}

func computeDigests(files []string, kind, seed uint) {
	var (
		digest  hash.Hash
		pattern string
		size    int
	)
	switch kind {
	case 0, 64:
		digest, pattern = xxh.New64(uint64(seed)), "%016x  %s - %dKB (%s)\n"
		size = 32 << 10
	case 32:
		digest, pattern = xxh.New32(uint32(seed)), "%08x %s - %dKB (%s)\n"
		size = 16 << 10
	default:
		return
	}

	buffer := make([]byte, size)
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
