package util

import (
	"math"
	"testing"
)

func TestMaxInt(t *testing.T) {
	testCases := []struct {
		x, y, expected int
	}{
		{0, 0, 0},
		{0, 10, 10},
		{10, 0, 10},
		{5, -10, 5},
		{-5, 10, 10},
		{math.MaxInt32, math.MinInt32, math.MaxInt32},
	}
	for _, test := range testCases {
		got := MaxInt(test.x, test.y)
		if got != test.expected {
			t.Errorf("MaxInt(%d, %d) should be %d but got %d", test.x, test.y, test.expected, got)
		}
	}
}

func TestMinInt(t *testing.T) {
	testCases := []struct {
		x, y, expected int
	}{
		{0, 0, 0},
		{20, 10, 10},
		{10, 20, 10},
		{5, -10, -10},
		{-5, 10, -5},
		{math.MaxInt32, math.MinInt32, math.MinInt32},
	}
	for _, test := range testCases {
		got := MinInt(test.x, test.y)
		if got != test.expected {
			t.Errorf("MinInt(%d, %d) should be %d but got %d", test.x, test.y, test.expected, got)
		}
	}
}

func TestMaxInt64(t *testing.T) {
	testCases := []struct {
		x, y, expected int64
	}{
		{0, 0, 0},
		{0, 10, 10},
		{10, 0, 10},
		{5, -10, 5},
		{-5, 10, 10},
		{math.MaxInt64, math.MinInt64, math.MaxInt64},
	}
	for _, test := range testCases {
		got := MaxInt64(test.x, test.y)
		if got != test.expected {
			t.Errorf("MaxInt64(%d, %d) should be %d but got %d", test.x, test.y, test.expected, got)
		}
	}
}

func TestMinInt64(t *testing.T) {
	testCases := []struct {
		x, y, expected int64
	}{
		{0, 0, 0},
		{20, 10, 10},
		{10, 20, 10},
		{5, -10, -10},
		{-5, 10, -5},
		{math.MaxInt64, math.MinInt64, math.MinInt64},
	}
	for _, test := range testCases {
		got := MinInt64(test.x, test.y)
		if got != test.expected {
			t.Errorf("MinInt64(%d, %d) should be %d but got %d", test.x, test.y, test.expected, got)
		}
	}
}
