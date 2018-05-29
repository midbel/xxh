package xxh

import (
	"bytes"
	"encoding/binary"
	"math/bits"
)

const (
	PRIME32_1 uint32 = 2654435761
	PRIME32_2        = 2246822519
	PRIME32_3        = 3266489917
	PRIME32_4        = 668265263
	PRIME32_5        = 374761393
)

func XXH32(bs []byte, seed uint32) uint32 {
	var (
		acc    uint32
		reader *bytes.Reader
	)
	if len(bs) >= 16 {
		reader, acc = accumulate32(bs, seed)
	} else {
		acc = seed + PRIME32_5
		reader = bytes.NewReader(bs)
	}
	acc += uint32(len(bs))
	for reader.Len() >= 4 {
		var v uint32
		binary.Read(reader, binary.LittleEndian, &v)
		acc = acc + (v * PRIME32_3)
		acc = bits.RotateLeft32(acc, 17) * PRIME32_4
	}
	for reader.Len() > 0 {
		v, _ := reader.ReadByte()
		acc = acc + uint32(v)*PRIME32_5
		acc = bits.RotateLeft32(acc, 11) * PRIME32_1
	}

	acc = acc ^ (acc >> 15)
	acc *= PRIME32_2
	acc = acc ^ (acc >> 13)
	acc *= PRIME32_3
	acc = acc ^ (acc >> 16)

	return acc
}

func accumulate32(bs []byte, seed uint32) (*bytes.Reader, uint32) {
	acc1 := seed + PRIME32_1 + PRIME32_2
	acc2 := seed + PRIME32_2
	acc3 := seed + 0
	acc4 := seed - PRIME32_1

	r := bytes.NewReader(bs)
	as := []uint32{acc1, acc2, acc3, acc4}

	var acc uint32
	for {
		bs := make([]byte, 16)
		n, _ := r.Read(bs)
		if n == len(bs) {
			as = processStripe(bs, as)
		} else {
			r = bytes.NewReader(bs[:n])
			break
		}
	}
	offsets := []int{1, 7, 12, 18}
	for i := range as {
		acc += bits.RotateLeft32(as[i], offsets[i])
	}
	return r, acc
}

func processStripe(bs []byte, as []uint32) []uint32 {
	r := bytes.NewReader(bs)
	for i := 0; r.Len() > 0; i++ {
		var v uint32
		binary.Read(r, binary.LittleEndian, &v)

		a := as[i%4] + (v * PRIME32_2)
		a = bits.RotateLeft32(a, 13)

		as[i] = a * PRIME32_1
	}
	return as
}
