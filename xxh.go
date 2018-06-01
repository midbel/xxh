package xxh

type Seeder interface {
  Seed(uint)
}

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

const (
	PRIME64_1 uint64 = 11400714785074694791
	PRIME64_2        = 14029467366897019727
	PRIME64_3        = 1609587929392839161
	PRIME64_4        = 9650029242287828579
	PRIME64_5        = 2870177450012600261
)

const (
  Block64 = 32
  Size64 = 8
)

var ints = []int{1, 7, 12, 18}
