package searcher

func patternToTarget(pattern []byte) []byte {
	var escape bool
	bs := make([]byte, 0, len(pattern))
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		if escape {
			switch b {
			case 'a':
				bs = append(bs, '\a')
			case 'b':
				bs = append(bs, '\b')
			case 'f':
				bs = append(bs, '\f')
			case 'n':
				bs = append(bs, '\n')
			case 'r':
				bs = append(bs, '\r')
			case 't':
				bs = append(bs, '\t')
			case 'v':
				bs = append(bs, '\v')
			case 'x':
				if i+2 < len(pattern) && isHex(pattern[i+1]) && isHex(pattern[i+2]) {
					bs = append(bs, hexToDigit(pattern[i+1])<<4|hexToDigit(pattern[i+2]))
					i += 2
				} else {
					bs = append(bs, b)
				}
			case 'v':
				bs = append(bs, '\v')
			default:
				bs = append(bs, b)
			}
			escape = false
		} else if b == '\\' {
			escape = true
		} else {
			bs = append(bs, b)
		}
	}
	if escape {
		bs = append(bs, '\\')
	}
	return bs
}

func isHex(b byte) bool {
	return '0' <= b && b <= '9' || 'A' <= b && b <= 'F' || 'a' <= b && b <= 'f'
}

func hexToDigit(b byte) byte {
	switch {
	case '0' <= b && b <= '9':
		return b - '0'
	case 'A' <= b && b <= 'F':
		return b - 'A' + 10
	case 'a' <= b && b <= 'f':
		return b - 'a' + 10
	default:
		return 0
	}
}
