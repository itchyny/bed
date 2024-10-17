package buffer

type constReader byte

// Read implements the io.Reader interface.
func (r constReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = byte(r)
	}
	return len(b), nil
}

// Seek implements the io.Seeker interface.
func (constReader) Seek(int64, int) (int64, error) {
	return 0, nil
}

// ReadAt implements the io.ReaderAt interface.
func (r constReader) ReadAt(b []byte, _ int64) (int, error) {
	return r.Read(b)
}
