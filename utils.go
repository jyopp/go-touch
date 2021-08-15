package fbui

// bytesFill optimally fills a byte array with a repeating pattern
func bytesFill(b, pattern []byte) {
	if l := copy(b, pattern); l > 0 {
		// Copy l*2^n bytes on each pass
		for ; l < len(b); l *= 2 {
			copy(b[l:], b[:l])
		}
	}
}
