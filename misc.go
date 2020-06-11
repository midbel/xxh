package xxh

import (
  "encoding/binary"
)

func appendUint8(bs []byte, x uint8) []byte {
  return append(bs, byte(x))
}

func consumeUint8(bs []byte) ([]byte, uint8) {
  return bs[1:], uint8(bs[0])
}

func appendUint32(bs []byte, x uint32) []byte {
  var a [4]byte
  binary.BigEndian.PutUint32(a[:], x)
  return append(bs, a[:]...)
}

func consumeUint32(bs []byte) ([]byte, uint32) {
  return bs[4:], binary.BigEndian.Uint32(bs)
}

func appendUint64(bs []byte, x uint64) []byte {
  var a [8]byte
  binary.BigEndian.PutUint64(a[:], x)
  return append(bs, a[:]...)
}

func consumeUint64(bs []byte) ([]byte, uint64) {
  return bs[8:], binary.BigEndian.Uint64(bs)
}
