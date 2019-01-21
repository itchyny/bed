package searcher

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/itchyny/bed/mathutil"
)

// Searcher represents a searcher.
type Searcher struct {
	r       readAtSeeker
	loopCh  chan struct{}
	cursor  int64
	pattern string
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// NewSearcher creates a new searcher.
func NewSearcher(r readAtSeeker) *Searcher {
	return &Searcher{r: r}
}

var errNotFound = errors.New("pattern not found")

const loadSize = 1024 * 1024

// Forward searches the pattern forward.
func (s *Searcher) Forward(cursor int64, pattern string) <-chan int64 {
	s.cursor, s.pattern = cursor, pattern
	ch := make(chan int64)
	s.loop(s.forward, ch)
	return ch
}

func (s *Searcher) forward() (int64, error) {
	target := []byte(s.pattern)
	base := s.cursor + 1
	n, bs, err := s.readBytes(base, loadSize)
	if err != nil {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound
	}
	s.cursor += int64(n)
	i := bytes.Index(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

// Backward searches the pattern forward.
func (s *Searcher) Backward(cursor int64, pattern string) <-chan int64 {
	s.cursor, s.pattern = cursor, pattern
	ch := make(chan int64)
	s.loop(s.backward, ch)
	return ch
}

func (s *Searcher) backward() (int64, error) {
	target := []byte(s.pattern)
	base := mathutil.MaxInt64(0, s.cursor-int64(loadSize))
	n, bs, err := s.readBytes(base, int(mathutil.MinInt64(int64(loadSize), s.cursor)))
	if err != nil {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound
	}
	s.cursor = base
	i := bytes.LastIndex(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

func (s *Searcher) loop(f func() (int64, error), ch chan<- int64) {
	if s.loopCh != nil {
		close(s.loopCh)
	}
	s.loopCh = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.loopCh:
				return
			case <-time.After(10 * time.Millisecond):
				idx, err := f()
				if err != nil {
					ch <- -1
					close(ch)
					return
				}
				if idx >= 0 {
					ch <- idx
					close(ch)
					return
				}
			}
		}
	}()
}

func (s *Searcher) readBytes(offset int64, len int) (int, []byte, error) {
	bytes := make([]byte, len)
	n, err := s.r.ReadAt(bytes, offset)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}
