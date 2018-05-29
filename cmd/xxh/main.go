package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/midbel/xxh"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	var group errgroup.Group
	for _, f := range flag.Args() {
		file := f
		group.Go(func() error {
			bs, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			log.Printf("%x %s", xxh.XXH32(bs, 0), file)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Println(err)
	}
}
