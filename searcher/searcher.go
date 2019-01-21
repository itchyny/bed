package searcher

import (
	"bytes"
	"fmt"
	"io"

	"github.com/itchyny/bed/mathutil"
)

// Searcher represents a searcher.
type Searcher struct {
	r readAtSeeker
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// NewSearcher creates a new searcher.
func NewSearcher(r readAtSeeker) *Searcher {
	return &Searcher{r: r}
}

// Forward searches the pattern forward.
func (s *Searcher) Forward(cursor int64, pattern string) (int64, error) {
	target := []byte(pattern)
	base, size := cursor+1, 1000
	_, bs, err := s.readBytes(base, size)
	if err != nil {
		return -1, err
	}
	i := bytes.Index(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, fmt.Errorf("pattern not found: %q", pattern)
}

// Backward searches the pattern backward.
func (s *Searcher) Backward(cursor int64, pattern string) (int64, error) {
	target := []byte(pattern)
	size := 1000
	base := mathutil.MaxInt64(0, cursor-int64(size))
	_, bs, err := s.readBytes(base, int(mathutil.MinInt64(int64(size), cursor)))
	if err != nil {
		return -1, err
	}
	i := bytes.LastIndex(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, fmt.Errorf("pattern not found: %q", pattern)
}

func (s *Searcher) readBytes(offset int64, len int) (int, []byte, error) {
	bytes := make([]byte, len)
	n, err := s.r.ReadAt(bytes, offset)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}
