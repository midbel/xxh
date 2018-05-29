package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/midbel/xxh"
	"golang.org/x/sync/errgroup"
)

func init() {
	log.SetFlags(0)
}

func main() {

	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()

	var g errgroup.Group
	for _, f := range flag.Args() {
		file := f
		g.Go(func() error {
			r, err := os.Open(file)
			if err != nil {
				return err
			}
			defer r.Close()

			w := xxh.New32(uint32(*seed))
			if _, err := io.Copy(w, r); err != nil {
				return err
			}
			log.Printf("%x %s", w.Sum32(), file)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Println(err)
	}
}
