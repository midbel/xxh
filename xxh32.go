package xxh

import (
	"encoding/binary"
	"fmt"
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

const (
	magic32 = "xxhash32\x03\x02"
	len32   = len(magic32) + sizeBlock32 + 1 + (6 * 4)
)

var default32 = New32(0)

type xxhash32 struct {
	size uint32
	seed uint32
	as   [4]uint32

	offset int
	buffer [sizeBlock32]byte
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

func (x *xxhash32) MarshalBinary() ([]byte, error) {
	bs := make([]byte, len32)
	bs = append(bs, magic32...)
	bs = append(bs, appendUint32(bs, x.size)...)
	bs = append(bs, appendUint32(bs, x.seed)...)
	bs = append(bs, appendUint32(bs, x.as[0])...)
	bs = append(bs, appendUint32(bs, x.as[1])...)
	bs = append(bs, appendUint32(bs, x.as[2])...)
	bs = append(bs, appendUint32(bs, x.as[3])...)
	bs = append(bs, appendUint8(bs, uint8(x.offset))...)
	if x.offset > 0 {
		bs = append(bs, x.buffer[:x.offset]...)
	}
	return bs, nil
}

func (x *xxhash32) UnmarshalBinary(bs []byte) error {
	if len(bs) < len(magic32) && string(bs[:len(magic32)]) != magic32 {
		return fmt.Errorf("invalid hash state identifier")
	}
	if len(bs) != len32 {
		return fmt.Errorf("invalid hash state size")
	}
	bs, x.size = consumeUint32(bs)
	bs, x.seed = consumeUint32(bs)

	bs, x.as[0] = consumeUint32(bs)
	bs, x.as[1] = consumeUint32(bs)
	bs, x.as[2] = consumeUint32(bs)
	bs, x.as[3] = consumeUint32(bs)

	if bs, offset := consumeUint8(bs); offset > 0 {
		x.offset = copy(x.buffer[:], bs[:x.offset])
	}
	return nil
	return nil
}

func (x *xxhash32) Size() int      { return sizeHash32 }
func (x *xxhash32) BlockSize() int { return sizeBlock32 }

func (x *xxhash32) Write(bs []byte) (int, error) {
	var i int
	if x.offset > 0 {
		i = copy(x.buffer[x.offset:], bs)
		x.offset += i
		if x.offset >= sizeBlock32 {
			x.offset = 0
			x.calculateBlock(x.buffer[:])
		}
	}
	size := len(bs)
	for i < size {
		if size-i < sizeBlock32 {
			break
		}
		x.calculateBlock(bs[i:])
		i += sizeBlock32
	}
	if diff := len(bs) - i; diff > 0 {
		x.offset = copy(x.buffer[:], bs[i:])
	}

	return size, nil
}

func (x *xxhash32) Seed(s uint) {
	x.seed = uint32(s)
	x.Reset()
}

func (x *xxhash32) Reset() {
	x.offset = 0
	x.as, x.size = reset32(x.seed), 0
}

func (x *xxhash32) Sum(bs []byte) []byte {
	y := *x
	hash := y.checksum()
	return append(bs, hash...)
}

func (x *xxhash32) checksum() []byte {
	var (
		acc    uint32
		buffer []byte
	)
	if x.offset > 0 {
		buffer = append(buffer, x.buffer[:x.offset]...)
	}
	if x.size == 0 {
		acc = x.seed + PRIME32_5
	} else {
		acc += bits.RotateLeft32(x.as[0], 1)
		acc += bits.RotateLeft32(x.as[1], 7)
		acc += bits.RotateLeft32(x.as[2], 12)
		acc += bits.RotateLeft32(x.as[3], 18)
	}
	z := len(buffer)
	acc += x.size + uint32(z)

	var i int
	for i < (z-sizeHash32)+1 {
		v := binary.LittleEndian.Uint32(buffer[i:]) * PRIME32_3
		acc += v
		acc = bits.RotateLeft32(acc, 17) * PRIME32_4

		i += sizeHash32
	}
	for i < z {
		acc += uint32(buffer[i]) * PRIME32_5
		acc = bits.RotateLeft32(acc, 11) * PRIME32_1

		i++
	}

	acc = (acc ^ (acc >> 15)) * PRIME32_2
	acc = (acc ^ (acc >> 13)) * PRIME32_3
	acc = acc ^ (acc >> 16)

	return []byte{
		byte(acc >> 24),
		byte((acc >> 16) & 0xFF),
		byte((acc >> 8) & 0xFF),
		byte(acc & 0xFF),
	}
}

func (x *xxhash32) Sum32() uint32 {
	return binary.BigEndian.Uint32(x.checksum())
}

func (x *xxhash32) calculateBlock(buf []byte) {
	x.as[0] = round32(x.as[0], binary.LittleEndian.Uint32(buf[0:]))
	x.as[1] = round32(x.as[1], binary.LittleEndian.Uint32(buf[4:]))
	x.as[2] = round32(x.as[2], binary.LittleEndian.Uint32(buf[8:]))
	x.as[3] = round32(x.as[3], binary.LittleEndian.Uint32(buf[12:]))

	x.size += sizeBlock32
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
