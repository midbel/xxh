package xxh

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/bits"
)

const (
	PRIME64_1 uint64 = 11400714785074694791
	PRIME64_2        = 14029467366897019727
	PRIME64_3        = 1609587929392839161
	PRIME64_4        = 9650029242287828579
	PRIME64_5        = 2870177450012600261
)

const (
	sizeHash64  = 8
	sizeBlock64 = 32
)

type xxhash64 struct {
	size   uint64
	seed   uint64
	as     [4]uint64
	buffer []byte
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
	x.Reset()

	return &x
}

func (x *xxhash64) Size() int      { return sizeHash64 }
func (x *xxhash64) BlockSize() int { return sizeBlock64 }

func (x *xxhash64) Write(bs []byte) (int, error) {
	var n int
	for i := 0; i < len(bs); i += sizeBlock64 {
		if len(bs[i:]) < sizeBlock64 {
			n = i
			break
		}
		x.calculateBlock(bs[i:])
	}
	x.buffer = append(x.buffer, bs[n:]...)

	return len(bs), nil
}

func (x *xxhash64) Seed(s uint) {
	x.Reset()
	x.seed = uint64(s)
}

func (x *xxhash64) Reset() {
	x.buffer = x.buffer[:0]
	x.as, x.size = reset64(x.seed), 0
}

func (x *xxhash64) Sum(bs []byte) []byte {
	var acc uint64

	x.buffer = append(x.buffer, bs...)
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
	acc += x.size + uint64(len(x.buffer))

	var i int
	z := len(x.buffer)
	for i = 0; i < z-sizeHash64; i += sizeHash64 {
		v := binary.LittleEndian.Uint64(x.buffer[i:])
		acc = acc ^ round64(0, v)
		acc = bits.RotateLeft64(acc, 27) * PRIME64_1
		acc += PRIME64_4
	}
	if (z - i) >= 4 {
		v := binary.LittleEndian.Uint32(x.buffer[i:])
		acc = acc ^ (uint64(v) * PRIME64_1)
		acc = (bits.RotateLeft64(acc, 23)) * PRIME64_2
		acc += PRIME64_3
	}
	for ; i < z; i++ {
		acc = acc ^ uint64(x.buffer[i])*PRIME64_5
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

func (x *xxhash64) calculateBlock(buffer []byte) {
	for j := 0; j < 4; j++ {
		v := binary.LittleEndian.Uint64(buffer[j*sizeHash64:])
		a := x.as[j] + (v * PRIME64_2)
		a = bits.RotateLeft64(a, 31)

		x.as[j] = a * PRIME64_1
	}
	x.size += uint64(x.BlockSize())
}

func (x *xxhash64) calculate() {
	z := x.BlockSize()
	for i := 0; i < len(x.buffer); i += z {
		if len(x.buffer[i:]) < z {
			x.buffer = x.buffer[i:]
			return
		}
		x.calculateBlock(x.buffer[i:])
		// for j := 0; j < 4; j++ {
		// 	v := binary.LittleEndian.Uint64(x.buffer[i+(j*8):])
		// 	a := x.as[j] + (v * PRIME64_2)
		// 	a = bits.RotateLeft64(a, 31)
		//
		// 	x.as[j] = a * PRIME64_1
		// }
		// x.size += uint64(z)
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
