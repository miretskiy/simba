package algo

// Each64 invokes fn for every full 64-byte chunk in b and returns the tail.
// It is designed to be inlined and incur zero overhead.
func Each64(b []byte, fn func(chunk []byte)) (tail []byte) {
	for len(b) >= 64 {
		fn(b[:64])
		b = b[64:]
	}
	return b
}

// Each32 iterates over 32-byte chunks.
func Each32(b []byte, fn func(chunk []byte)) (tail []byte) {
	for len(b) >= 32 {
		fn(b[:32])
		b = b[32:]
	}
	return b
}

// Each16 iterates over 16-byte chunks.
func Each16(b []byte, fn func(chunk []byte)) (tail []byte) {
	for len(b) >= 16 {
		fn(b[:16])
		b = b[16:]
	}
	return b
}
