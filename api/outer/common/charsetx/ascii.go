package charsetx

func IsASCIIChar(r rune) bool {
	if r < 0 || r > 127 {
		return false
	}

	return true
}
