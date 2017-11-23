package util

func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func MinInt(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func MinInt64(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}
