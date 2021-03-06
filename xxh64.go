package xxh

import (
	"encoding/binary"
	"fmt"
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

const (
	magic64 = "xxhash64\x06\x04"
	len64   = len(magic64) + sizeBlock64 + 1 + (6 * 8)
)

var default64 = New64(0)

type xxhash64 struct {
	size uint64
	seed uint64
	as   [4]uint64

	offset int
	buffer [sizeBlock64]byte
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

func (x *xxhash64) MarshalBinary() ([]byte, error) {
	bs := make([]byte, 0, len64)
	bs = append(bs, magic64...)
	bs = append(bs, appendUint64(bs, x.size)...)
	bs = append(bs, appendUint64(bs, x.seed)...)
	bs = append(bs, appendUint64(bs, x.as[0])...)
	bs = append(bs, appendUint64(bs, x.as[1])...)
	bs = append(bs, appendUint64(bs, x.as[2])...)
	bs = append(bs, appendUint64(bs, x.as[3])...)
	bs = append(bs, appendUint8(bs, uint8(x.offset))...)
	if x.offset > 0 {
		bs = append(bs, x.buffer[:x.offset]...)
	}
	return bs, nil
}

func (x *xxhash64) UnmarshalBinary(bs []byte) error {
	if len(bs) < len(magic64) && string(bs[:len(magic64)]) != magic64 {
		return fmt.Errorf("invalid hash state identifier")
	}
	if len(bs) != len64 {
		return fmt.Errorf("invalid hash state size")
	}
	bs, x.size = consumeUint64(bs)
	bs, x.seed = consumeUint64(bs)

	bs, x.as[0] = consumeUint64(bs)
	bs, x.as[1] = consumeUint64(bs)
	bs, x.as[2] = consumeUint64(bs)
	bs, x.as[3] = consumeUint64(bs)

	if bs, offset := consumeUint8(bs); offset > 0 {
		x.offset = copy(x.buffer[:], bs[:x.offset])
	}
	return nil
}

func (x *xxhash64) Size() int      { return sizeHash64 }
func (x *xxhash64) BlockSize() int { return sizeBlock64 }

func (x *xxhash64) Write(bs []byte) (int, error) {
	var i int
	if x.offset > 0 {
		i = copy(x.buffer[x.offset:], bs)
		x.offset += i
		if x.offset >= sizeBlock64 {
			x.offset = 0
			x.calculateBlock(x.buffer[:])
		}
	}
	size := len(bs)
	for i < size {
		if size-i < sizeBlock64 {
			break
		}
		x.calculateBlock(bs[i:])
		i += sizeBlock64
	}
	if diff := len(bs) - i; diff > 0 {
		x.offset = copy(x.buffer[:], bs[i:])
	}

	return size, nil
}

func (x *xxhash64) Seed(s uint) {
	x.seed = uint64(s)
	x.Reset()
}

func (x *xxhash64) Reset() {
	x.offset = 0
	x.as, x.size = reset64(x.seed), 0
}

func (x *xxhash64) Sum(bs []byte) []byte {
	y := *x
	hash := y.checksum()
	return append(bs, hash...)
}

func (x *xxhash64) checksum() []byte {
	var (
		acc    uint64
		buffer []byte
	)
	if x.offset > 0 {
		buffer = append(buffer, x.buffer[:x.offset]...)
	}
	if x.size == 0 && len(buffer) < sizeBlock64 {
		acc = x.seed + PRIME64_5
	} else {
		acc += bits.RotateLeft64(x.as[0], 1)
		acc += bits.RotateLeft64(x.as[1], 7)
		acc += bits.RotateLeft64(x.as[2], 12)
		acc += bits.RotateLeft64(x.as[3], 18)

		acc = merge64(acc, x.as[0])
		acc = merge64(acc, x.as[1])
		acc = merge64(acc, x.as[2])
		acc = merge64(acc, x.as[3])
	}
	z := len(buffer)
	acc += x.size + uint64(z)

	var i int
	for i < (z-sizeHash64)+1 {
		v := binary.LittleEndian.Uint64(buffer[i:])
		acc = acc ^ round64(0, v)
		acc = bits.RotateLeft64(acc, 27) * PRIME64_1
		acc += PRIME64_4

		i += sizeHash64
	}
	if (z - i) >= 4 {
		v := binary.LittleEndian.Uint32(buffer[i:])
		acc = acc ^ (uint64(v) * PRIME64_1)
		acc = bits.RotateLeft64(acc, 23) * PRIME64_2
		acc += PRIME64_3

		i += 4
	}
	for i < z {
		acc = acc ^ (uint64(buffer[i]) * PRIME64_5)
		acc = bits.RotateLeft64(acc, 11) * PRIME64_1

		i++
	}

	acc = acc ^ (acc >> 33)
	acc *= PRIME64_2
	acc = acc ^ (acc >> 29)
	acc *= PRIME64_3
	acc = acc ^ (acc >> 32)

	return []byte{
		byte(acc >> 56),
		byte((acc >> 48) & 0xFF),
		byte((acc >> 40) & 0xFF),
		byte((acc >> 32) & 0xFF),
		byte((acc >> 24) & 0xFF),
		byte((acc >> 16) & 0xFF),
		byte((acc >> 8) & 0xFF),
		byte(acc & 0xFF),
	}
}

func (x *xxhash64) Sum64() uint64 {
	return binary.BigEndian.Uint64(x.checksum())
}

func (x *xxhash64) calculateBlock(buf []byte) {
	x.as[0] = round64(x.as[0], binary.LittleEndian.Uint64(buf[0:]))
	x.as[1] = round64(x.as[1], binary.LittleEndian.Uint64(buf[8:]))
	x.as[2] = round64(x.as[2], binary.LittleEndian.Uint64(buf[16:]))
	x.as[3] = round64(x.as[3], binary.LittleEndian.Uint64(buf[24:]))

	x.size += sizeBlock64
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
