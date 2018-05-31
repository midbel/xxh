package xxh

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/bits"
	// "log"
)

const (
	PRIME32_1 uint32 = 2654435761
	PRIME32_2        = 2246822519
	PRIME32_3        = 3266489917
	PRIME32_4        = 668265263
	PRIME32_5        = 374761393
)

const (
	Block32 = 16
	Size32  = 4
)

type xxhash32 struct {
	size   uint32
	seed   uint32
	as     [4]uint32
	buffer *bytes.Buffer
}

func Sum32(bs []byte, seed uint32) uint32 {
	w := New32(seed)
	if _, err := io.Copy(w, bytes.NewReader(bs)); err != nil {
		return 0
	}
	return w.Sum32()
}

func New32(seed uint32) hash.Hash32 {
	var x xxhash32
	x.seed = seed
	x.buffer = new(bytes.Buffer)
	x.Reset()

	return &x
}

func (x *xxhash32) Size() int      { return Size32 }
func (x *xxhash32) BlockSize() int { return Block32 }

func (x *xxhash32) Write(bs []byte) (int, error) {
	x.buffer.Write(bs)

	x.calculate()

	return len(bs), nil
}

func (x *xxhash32) Reset() {
	x.buffer.Reset()
	x.as, x.size = reset32(x.seed), 0
}

func (x *xxhash32) Sum(bs []byte) []byte {
	var acc uint32

	x.buffer.Write(bs)
	if x.size == 0 {
		acc = x.seed + PRIME32_5
	} else {
		x.calculate()
		ix := []int{1, 7, 12, 18}
		for i := range x.as {
			acc += bits.RotateLeft32(x.as[i], ix[i])
		}
	}
	acc += x.size + uint32(x.buffer.Len())

	bs = x.buffer.Bytes()
	for len(bs) >= Size32 {
		v := binary.LittleEndian.Uint32(bs[:Size32])
		acc = acc + (v * PRIME32_3)
		acc = bits.RotateLeft32(acc, 17) * PRIME32_4
		bs = bs[Size32:]
	}
	for i := 0; i < len(bs); i++ {
		acc = acc + uint32(bs[i])*PRIME32_5
		acc = bits.RotateLeft32(acc, 11) * PRIME32_1
	}

	acc = acc ^ (acc >> 15)
	acc *= PRIME32_2
	acc = acc ^ (acc >> 13)
	acc *= PRIME32_3
	acc = acc ^ (acc >> 16)

	cs := make([]byte, Size32)
	binary.BigEndian.PutUint32(cs, acc)
	return cs
}

func (x *xxhash32) Sum32() uint32 {
	bs := x.Sum(nil)
	return binary.BigEndian.Uint32(bs)
}

func (x *xxhash32) calculate() {
	for {
		bs := make([]byte, Block32)
		if n, _ := io.ReadFull(x.buffer, bs); n < Block32 {
			x.buffer.Write(bs[:n])
			break
		}
		for i, j := 0, 0; i < len(bs); i, j = i+Size32, j+1 {
			v := binary.LittleEndian.Uint32(bs[i:i+Size32])
			a := x.as[j] + (v * PRIME32_2)
			a = bits.RotateLeft32(a, 13)

			x.as[j] = a * PRIME32_1
		}
		x.size += uint32(len(bs))
	}
}

func reset32(seed uint32) [4]uint32 {
	acc1 := seed + PRIME32_1 + PRIME32_2
	acc2 := seed + PRIME32_2
	acc3 := seed + 0
	acc4 := seed - PRIME32_1

	return [4]uint32{acc1, acc2, acc3, acc4}
}
