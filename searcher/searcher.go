package searcher

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/itchyny/bed/mathutil"
)

// Searcher represents a searcher.
type Searcher struct {
	r       readAtSeeker
	loopCh  chan struct{}
	cursor  int64
	pattern string
	mu      *sync.Mutex
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

// NewSearcher creates a new searcher.
func NewSearcher(r readAtSeeker) *Searcher {
	return &Searcher{r: r, mu: new(sync.Mutex)}
}

type errNotFound string

func (err errNotFound) Error() string {
	return "pattern not found: " + string(err)
}

const loadSize = 1024 * 1024

// Search the pattern.
func (s *Searcher) Search(cursor int64, pattern string, forward bool) <-chan interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursor, s.pattern = cursor, pattern
	ch := make(chan interface{})
	if forward {
		s.loop(s.forward, ch)
	} else {
		s.loop(s.backward, ch)
	}
	return ch
}

func (s *Searcher) forward() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	target := []byte(s.pattern)
	base := s.cursor + 1
	n, bs, err := s.readBytes(base, loadSize)
	if err != nil {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound(s.pattern)
	}
	s.cursor += int64(n)
	i := bytes.Index(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

func (s *Searcher) backward() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	target := []byte(s.pattern)
	base := mathutil.MaxInt64(0, s.cursor-int64(loadSize))
	n, bs, err := s.readBytes(base, int(mathutil.MinInt64(int64(loadSize), s.cursor)))
	if err != nil {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound(s.pattern)
	}
	s.cursor = base
	i := bytes.LastIndex(bs, target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

func (s *Searcher) loop(f func() (int64, error), ch chan<- interface{}) {
	if s.loopCh != nil {
		close(s.loopCh)
	}
	loopCh := make(chan struct{})
	s.loopCh = loopCh
	go func() {
		defer close(ch)
		for {
			select {
			case <-loopCh:
				return
			case <-time.After(10 * time.Millisecond):
				idx, err := f()
				if err != nil {
					ch <- err
					return
				}
				if idx >= 0 {
					ch <- idx
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
