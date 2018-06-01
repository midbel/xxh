package xxh

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/bits"
)

type xxhash32 struct {
	size   uint32
	seed   uint32
	as     [4]uint32
	buffer []byte
}

func Sum32(bs []byte, seed uint32) uint32 {
	w := New32(seed)
	if _, err := io.Copy(w, bytes.NewReader(bs)); err != nil {
		return 0
	}
	return w.Sum32()
}

func New32(seed uint32) hash.Hash32 {
	x := xxhash32{seed: seed}
	x.Reset()

	return &x
}

func (x *xxhash32) Size() int      { return Size32 }
func (x *xxhash32) BlockSize() int { return Block32 }

func (x *xxhash32) Write(bs []byte) (int, error) {
	x.buffer = append(x.buffer, bs...)
	x.calculate()
	return len(bs), nil
}

func (x *xxhash32) Seed(s uint) {
	x.Reset()
	x.seed = uint32(s)
}

func (x *xxhash32) Reset() {
	x.buffer = nil
	x.as, x.size = reset32(x.seed), 0
}

func (x *xxhash32) Sum(bs []byte) []byte {
	if len(bs) > 0 {
		x.buffer = append(x.buffer, bs...)
	}

	var acc uint32
	if x.size == 0 {
		acc = x.seed + PRIME32_5
	} else {
		x.calculate()
		for i := range x.as {
			acc += bits.RotateLeft32(x.as[i], ints[i])
		}
	}
	acc += x.size + uint32(len(x.buffer))

	for len(x.buffer) >= Size32 {
		acc += binary.LittleEndian.Uint32(x.buffer[:Size32]) * PRIME32_3
		acc = bits.RotateLeft32(acc, 17) * PRIME32_4
		x.buffer = x.buffer[Size32:]
	}
	for i := 0; i < len(x.buffer); i++ {
		acc += uint32(x.buffer[i]) * PRIME32_5
		acc = bits.RotateLeft32(acc, 11) * PRIME32_1
	}

	acc = (acc ^ (acc >> 15)) * PRIME32_2
	acc = (acc ^ (acc >> 13)) * PRIME32_3
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
	for n := len(x.buffer); n >= Block32; n -= Block32 {
		for i, j := 0, 0; i < Block32; i, j = i+Size32, j+1 {
			v := binary.LittleEndian.Uint32(x.buffer[i : i+Size32])
			a := x.as[j] + (v * PRIME32_2)

			x.as[j] = bits.RotateLeft32(a, 13) * PRIME32_1
		}
		x.buffer, x.size = x.buffer[Block32:], x.size+Block32
	}
}

func reset32(seed uint32) [4]uint32 {
	acc1 := seed + PRIME32_1 + PRIME32_2
	acc2 := seed + PRIME32_2
	acc3 := seed + 0
	acc4 := seed - PRIME32_1

	return [4]uint32{acc1, acc2, acc3, acc4}
}
