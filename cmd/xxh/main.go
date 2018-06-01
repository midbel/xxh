package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/midbel/xxh"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}

func main() {
	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()

	for _, f := range flag.Args() {
		r, err := os.Open(f)
		if err != nil {
			continue
		}
		defer r.Close()

		w := xxh.New32(uint32(*seed))
		if _, err := io.Copy(w, r); err != nil {
			continue
		}
		log.Printf("%x %s", w.Sum32(), f)
	}
}
