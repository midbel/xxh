package xxh

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/bits"
)

type xxhash64 struct {
	size   uint64
	seed   uint64
	as     [4]uint64
	buffer *bytes.Buffer
}

func Sum64(bs []byte, seed uint64) uint64 {
	w := New64(seed)
	if _, err := io.Copy(w, bytes.NewReader(bs)); err != nil {
		return 0
	}
	return w.Sum64()
}

func New64(seed uint64) hash.Hash64 {
	var x xxhash64
	x.seed = seed
	x.buffer = new(bytes.Buffer)
	x.Reset()

	return &x
}

func (x *xxhash64) Size() int      { return 8 }
func (x *xxhash64) BlockSize() int { return 32 }

func (x *xxhash64) Write(bs []byte) (int, error) {
	x.buffer.Write(bs)

	x.calculate()

	return len(bs), nil
}

func (x *xxhash64) Seed(s uint) {
	x.Reset()
	x.seed = uint64(s)
}

func (x *xxhash64) Reset() {
	x.buffer.Reset()
	x.as, x.size = reset64(x.seed), 0
}

func (x *xxhash64) Sum(bs []byte) []byte {
	var acc uint64

	x.buffer.Write(bs)
	if x.size == 0 {
		acc = x.seed + PRIME64_5
	} else {
		x.calculate()
		for i := range x.as {
			acc += bits.RotateLeft64(x.as[i], ints[i])
		}
		for i := range x.as {
			acc = merge64(acc, x.as[i])
		}
	}
	acc += x.size + uint64(x.buffer.Len())

	for x.buffer.Len() >= 8 {
		var v uint64
		binary.Read(x.buffer, binary.LittleEndian, &v)
		acc = acc ^ round64(0, v)
		acc = bits.RotateLeft64(acc, 27) * PRIME64_1
		acc += PRIME64_4
	}
	if x.buffer.Len() >= 4 {
		var v uint32
		binary.Read(x.buffer, binary.LittleEndian, &v)
		acc = acc ^ (uint64(v) * PRIME64_1)
		acc = (bits.RotateLeft64(acc, 23)) * PRIME64_2
		acc += PRIME64_3
	}
	for x.buffer.Len() > 0 {
		v, _ := x.buffer.ReadByte()
		acc = acc ^ uint64(v)*PRIME64_5
		acc = bits.RotateLeft64(acc, 11) * PRIME64_1
	}

	acc = acc ^ (acc >> 33)
	acc *= PRIME64_2
	acc = acc ^ (acc >> 29)
	acc *= PRIME64_3
	acc = acc ^ (acc >> 32)

	cs := make([]byte, x.Size())
	binary.BigEndian.PutUint64(cs, acc)
	return cs
}

func (x *xxhash64) Sum64() uint64 {
	bs := x.Sum(nil)
	return binary.BigEndian.Uint64(bs)
}

func (x *xxhash64) calculate() {
	for {
		bs := make([]byte, x.BlockSize())
		if n, _ := io.ReadFull(x.buffer, bs); n < x.BlockSize() {
			x.buffer.Write(bs[:n])
			break
		}
		r := bytes.NewReader(bs)
		for i := 0; r.Len() > 0; i++ {
			var v uint64
			binary.Read(r, binary.LittleEndian, &v)

			a := x.as[i%4] + (v * PRIME64_2)
			a = bits.RotateLeft64(a, 31)

			x.as[i] = a * PRIME64_1
		}
		x.size += uint64(len(bs))
	}
}

func reset64(seed uint64) [4]uint64 {
	acc1 := seed + PRIME64_1 + PRIME64_2
	acc2 := seed + PRIME64_2
	acc3 := seed + 0
	acc4 := seed - PRIME64_1

	return [4]uint64{acc1, acc2, acc3, acc4}
}

func merge64(a, curr uint64) uint64 {
	a = a ^ round64(0, curr)
	a = a * PRIME64_1
	return a + PRIME64_4
}

func round64(a, curr uint64) uint64 {
	a = a + (curr * PRIME64_2)
	a = bits.RotateLeft64(a, 31)
	return a * PRIME64_1
}
