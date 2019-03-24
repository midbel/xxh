package xxh

import (
	"encoding/binary"
	"hash"
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

var default64 = New64(0)

type xxhash64 struct {
	size   uint64
	seed   uint64
	as     [4]uint64
	buffer []byte
}

func Sum64(bs []byte, seed uint64) uint64 {
	default64.Reset()
	if _, err := default64.Write(bs); err != nil {
		return 0
	}
	return default64.Sum64()
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
	if len(x.buffer) > 0 {
		bs = append(x.buffer, bs...)
	}
	size := len(bs)
	var i int
	for i < size {
		if size - i < sizeBlock64 {
			break
		}
		x.calculateBlock(bs[i:])
		i += sizeBlock64
	}
	x.buffer = append(x.buffer[:0], bs[i:]...)

	return len(bs), nil
}

func (x *xxhash64) Seed(s uint) {
	x.seed = uint64(s)
	x.Reset()
}

func (x *xxhash64) Reset() {
	x.buffer = nil
	x.as, x.size = reset64(x.seed), 0
}

func (x *xxhash64) Sum(bs []byte) []byte {
	var acc uint64

	x.buffer = append(x.buffer, bs...)
	if x.size == 0 {
		acc = x.seed + PRIME64_5
	} else {
		if len(x.buffer) >= sizeBlock64 {
			x.calculate()
		}
		acc += bits.RotateLeft64(x.as[0], 1)
		acc += bits.RotateLeft64(x.as[1], 7)
		acc += bits.RotateLeft64(x.as[2], 12)
		acc += bits.RotateLeft64(x.as[3], 18)

		acc = merge64(acc, x.as[0])
		acc = merge64(acc, x.as[1])
		acc = merge64(acc, x.as[2])
		acc = merge64(acc, x.as[3])
	}
	z := len(x.buffer)
	acc += x.size + uint64(z)

	var i int
	for i = 0; i < z-sizeHash64; i += sizeHash64 {
		v := binary.LittleEndian.Uint64(x.buffer[i:])
		acc = acc ^ round64(0, v)
		acc = bits.RotateLeft64(acc, 27) * PRIME64_1
		acc += PRIME64_4
	}
	if (z - i) >= 4 {
		v := binary.LittleEndian.Uint32(x.buffer[i:])
		acc = acc ^ (uint64(v) * PRIME64_1)
		acc = bits.RotateLeft64(acc, 23) * PRIME64_2
		acc += PRIME64_3
		i+=4
	}
	for ; i < z; i++ {
		acc = acc ^ (uint64(x.buffer[i])*PRIME64_5)
		acc = bits.RotateLeft64(acc, 11) * PRIME64_1
	}

	acc = acc ^ (acc >> 33)
	acc *= PRIME64_2
	acc = acc ^ (acc >> 29)
	acc *= PRIME64_3
	acc = acc ^ (acc >> 32)

	x.buffer = nil

	cs := make([]byte, x.Size())
	binary.BigEndian.PutUint64(cs, acc)
	return cs
}

func (x *xxhash64) Sum64() uint64 {
	bs := x.Sum(nil)
	return binary.BigEndian.Uint64(bs)
}

func (x *xxhash64) calculateBlock(buf []byte) {
	x.as[0] = round64(x.as[0], binary.LittleEndian.Uint64(buf[0:]))
	x.as[1] = round64(x.as[1], binary.LittleEndian.Uint64(buf[8:]))
	x.as[2] = round64(x.as[2], binary.LittleEndian.Uint64(buf[16:]))
	x.as[3] = round64(x.as[3], binary.LittleEndian.Uint64(buf[24:]))
	// for j := 0; j < 4; j++ {
	// 	v := binary.LittleEndian.Uint64(buffer[j*sizeHash64:])
	// 	x.as[j] = round64(x.as[j], v)
	// }
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
	return (a * PRIME64_1) + PRIME64_4
}

func round64(a, curr uint64) uint64 {
	a += curr * PRIME64_2
	return bits.RotateLeft64(a, 31) * PRIME64_1
}
