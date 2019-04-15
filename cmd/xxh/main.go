package main

import (
	"bufio"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/midbel/xxh"
)

func main() {
	cmp := flag.Bool("c", false, "compare digest")
	kind := flag.Uint("k", 0, "hash")
	seed := flag.Uint("s", 0, "seed value")
	flag.Parse()

	var err error
	if *cmp {
		err = compareDigests(flag.Arg(0), *kind, *seed)
	} else {
		files, err := listfiles(flag.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = computeDigests(files, *kind, *seed)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

var buffer = make([]byte, 8<<10)

func listfiles(files []string) ([]string, error) {
	if len(files) > 0 {
		return files, nil
	}
	rs := bufio.NewScanner(bufio.NewReader(os.Stdin))
	for rs.Scan() {
		files = append(files, rs.Text())
	}
	return files, rs.Err()
}

func compareDigests(file string, kind, seed uint) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()

	var digest  hash.Hash
	switch kind {
	case 0, 64:
		digest = xxh.New64(uint64(seed))
	case 32:
		digest = xxh.New32(uint32(seed))
	default:
		return fmt.Errorf("unknown hash version")
	}
	rs := bufio.NewReader(r)

	var errcount int
	for {
		var want, file string
		n, err := fmt.Fscanf(rs, "%s  %s\n", &want, &file)
		if n != 2 && err == nil {
			return fmt.Errorf("not enough element scanned")
		}
		switch err {
		case nil:
			sum, err := calculate(file, digest)
			if err != nil {
				continue
			}
			if hs := fmt.Sprintf("%016x", sum); hs != want {
				errcount++
				fmt.Fprintf(os.Stderr, "%s: digest does not match\n", file)
			}
		case io.EOF:
			if errcount == 0 {
				fmt.Fprintln(os.Stdout, "OK")
			}
			return nil
		default:
			return err
		}
	}
}

func computeDigests(files []string, kind, seed uint) error {
	var (
		digest  hash.Hash
		pattern string
	)
	switch kind {
	case 0, 64:
		digest, pattern = xxh.New64(uint64(seed)), "%016x  %s\n"
	case 32:
		digest, pattern = xxh.New32(uint32(seed)), "%08x  %s\n"
	default:
		return fmt.Errorf("unknown hash version")
	}
	for i := 0; i < len(files); i++ {
		sum, err := calculate(files[i], digest)
		if err == nil {
			fmt.Printf(pattern, sum, files[i])
		}
	}
	return nil
}

func calculate(file string, digest hash.Hash) ([]byte, error) {
	defer digest.Reset()

	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	_, err = io.CopyBuffer(digest, r, buffer)
	return digest.Sum(nil), err
}
