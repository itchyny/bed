package mathutil

// MaxInt returns the larger of x or y.
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// MinInt returns the smaller of x or y.
func MinInt(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// MaxInt64 returns the larger of x or y.
func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// MinInt64 returns the smaller of x or y.
func MinInt64(x, y int64) int64 {
	if x > y {
		return y
	}
	return x
}
