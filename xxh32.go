package xxh

import (
	"encoding/binary"
	"hash"
	"math/bits"
)

const (
	PRIME32_1 uint32 = 2654435761
	PRIME32_2        = 2246822519
	PRIME32_3        = 3266489917
	PRIME32_4        = 668265263
	PRIME32_5        = 374761393
)

const (
	sizeBlock32 = 16
	sizeHash32  = 4
)

var default32 = New32(0)

type xxhash32 struct {
	size   uint32
	seed   uint32
	as     [4]uint32
	buffer []byte

	sum [sizeBlock32]byte
}

func Sum32(bs []byte, seed uint32) uint32 {
	default32.Reset()
	if _, err := default32.Write(bs); err != nil {
		return 0
	}
	return default32.Sum32()
}

func New32(seed uint32) hash.Hash32 {
	var x xxhash32
	x.seed = seed
	x.Reset()

	return &x
}

func (x *xxhash32) Size() int      { return sizeHash32 }
func (x *xxhash32) BlockSize() int { return sizeBlock32 }

func (x *xxhash32) Write(bs []byte) (int, error) {
	if len(x.buffer) > 0 {
		bs = append(x.buffer, bs...)
	}
	size := len(bs)
	var i int
	for i < size {
		if size-i < sizeBlock32 {
			break
		}
		x.calculateBlock(bs[i:])
		i += sizeBlock32
	}
	x.buffer = bs[i:]

	return size, nil
}

func (x *xxhash32) Seed(s uint) {
	x.seed = uint32(s)
	x.Reset()
}

func (x *xxhash32) Reset() {
	x.buffer = nil
	x.as, x.size = reset32(x.seed), 0
}

func (x *xxhash32) Sum(bs []byte) []byte {
	defer x.Reset()

	var acc uint32

	if len(bs) > 0 {
		x.buffer = append(x.buffer, bs...)
	}
	if x.size == 0 {
		acc = x.seed + PRIME32_5
	} else {
		acc += bits.RotateLeft32(x.as[0], 1)
		acc += bits.RotateLeft32(x.as[1], 7)
		acc += bits.RotateLeft32(x.as[2], 12)
		acc += bits.RotateLeft32(x.as[3], 18)
	}
	z := len(x.buffer)
	acc += x.size + uint32(z)

	var i int
	for i < (z-sizeHash32)+1 {
		v := binary.LittleEndian.Uint32(x.buffer[i:]) * PRIME32_3
		acc += v
		acc = bits.RotateLeft32(acc, 17) * PRIME32_4

		i += sizeHash32
	}
	for i < z {
		acc += uint32(x.buffer[i]) * PRIME32_5
		acc = bits.RotateLeft32(acc, 11) * PRIME32_1

		i++
	}

	acc = (acc ^ (acc >> 15)) * PRIME32_2
	acc = (acc ^ (acc >> 13)) * PRIME32_3
	acc = acc ^ (acc >> 16)

	// cs := make([]byte, sizeHash32)
	binary.BigEndian.PutUint32(x.sum[:], acc)
	return x.sum[:]
}

func (x *xxhash32) Sum32() uint32 {
	bs := x.Sum(nil)
	return binary.BigEndian.Uint32(bs)
}

func (x *xxhash32) calculateBlock(buf []byte) {
	x.as[0] = round32(x.as[0], binary.LittleEndian.Uint32(buf[0:]))
	x.as[1] = round32(x.as[1], binary.LittleEndian.Uint32(buf[4:]))
	x.as[2] = round32(x.as[2], binary.LittleEndian.Uint32(buf[8:]))
	x.as[3] = round32(x.as[3], binary.LittleEndian.Uint32(buf[12:]))

	x.size += sizeBlock32
}

func (x *xxhash32) calculate() {
	size := len(x.buffer)
	var i int
	for i < size {
		x.calculateBlock(x.buffer[i:])
		i += sizeBlock32
	}
	if i < size {
		x.buffer = x.buffer[i:]
	}
}

func round32(a, curr uint32) uint32 {
	a += curr * PRIME32_2
	return bits.RotateLeft32(a, 13) * PRIME32_1
}

func reset32(seed uint32) [4]uint32 {
	acc1 := seed + PRIME32_1 + PRIME32_2
	acc2 := seed + PRIME32_2
	acc3 := seed + 0
	acc4 := seed - PRIME32_1

	return [4]uint32{acc1, acc2, acc3, acc4}
}
