package endian

func Read16le(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func Read32le(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 |
		uint32(b[2])<<16 | uint32(b[3])<<24
}

func Read64le(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 |
		uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 |
		uint64(b[6])<<48 | uint64(b[7])<<56
}

func Put16le(b []byte, x uint16) {
	b[0] = byte(x)
	b[1] = byte(x >> 8)
}

func Put32le(b []byte, x uint32) {
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
}

func Put48le(b []byte, x uint64) {
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
	b[4] = byte(x >> 32)
	b[5] = byte(x >> 40)
}

func Put64le(b []byte, x uint64) {
	b[0] = byte(x)
	b[1] = byte(x >> 8)
	b[2] = byte(x >> 16)
	b[3] = byte(x >> 24)
	b[4] = byte(x >> 32)
	b[5] = byte(x >> 40)
	b[6] = byte(x >> 48)
	b[7] = byte(x >> 56)
}

func Swap64(x uint64) uint64 {
	return ((x & 0x00000000000000FF) << 56) |
		((x & 0x000000000000FF00) << 40) |
		((x & 0x0000000000FF0000) << 24) |
		((x & 0x00000000FF000000) << 8) |
		((x & 0x000000FF00000000) >> 8) |
		((x & 0x0000FF0000000000) >> 24) |
		((x & 0x00FF000000000000) >> 40) |
		((x & 0xFF00000000000000) >> 56)
}

func Swap48(x uint64) uint64 {
	return ((x & 0x000000000000FF00) << 40) |
		((x & 0x0000000000FF0000) << 24) |
		((x & 0x00000000FF000000) << 8) |
		((x & 0x000000FF00000000) >> 8) |
		((x & 0x0000FF0000000000) >> 24) |
		((x & 0x00FF000000000000) >> 40)
}

func Swap32(x uint32) uint32 {
	return (x&0x000000FF)<<24 | (x&0x0000FF00)<<8 |
		(x&0x00FF0000)>>8 | (x&0xFF000000)>>24
}

func Swap16(x uint16) uint16 {
	return (x&0x00FF)<<8 | (x&0xFF00)>>8
}
