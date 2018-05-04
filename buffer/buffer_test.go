package buffer

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestBufferEmpty(t *testing.T) {
	b := NewBuffer(strings.NewReader(""))

	p := make([]byte, 10)
	n, err := b.Read(p)
	if err != io.EOF {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 0 {
		t.Errorf("n should be 0 but got: %d", n)
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 0 {
		t.Errorf("l should be 0 but got: %d", l)
	}
}

func TestBuffer(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))

	p := make([]byte, 8)
	n, err := b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "01234567" {
		t.Errorf("p should be 01234567 but got: %s", string(p))
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 16 {
		t.Errorf("l should be 16 but got: %d", l)
	}

	_, err = b.Seek(4, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "456789ab" {
		t.Errorf("p should be 456789ab but got: %s", string(p))
	}

	_, err = b.Seek(-4, io.SeekCurrent)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "89abcdef" {
		t.Errorf("p should be 89abcdef but got: %s", string(p))
	}

	_, err = b.Seek(-4, io.SeekEnd)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 4 {
		t.Errorf("n should be 4 but got: %d", n)
	}
	if string(p) != "cdefcdef" {
		t.Errorf("p should be cdefcdef but got: %s", string(p))
	}

	n, err = b.ReadAt(p, 7)
	if err != nil {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "789abcde" {
		t.Errorf("p should be 789abcde but got: %s", string(p))
	}

	n, err = b.ReadAt(p, -1)
	if err == nil {
		t.Errorf("err should not be nil but got: %v", err)
	}
	if n != 0 {
		t.Errorf("n should be 0 but got: %d", n)
	}
}

func TestBufferClone(t *testing.T) {
	b0 := NewBuffer(strings.NewReader("0123456789abcdef"))
	b1 := b0.Clone()

	bufferEqual := func(b0 *Buffer, b1 *Buffer) bool {
		if b0.index != b1.index || len(b0.rrs) != len(b1.rrs) {
			return false
		}
		for i := 0; i < len(b0.rrs); i++ {
			if b0.rrs[i].min != b1.rrs[i].min || b0.rrs[i].max != b1.rrs[i].max ||
				b0.rrs[i].diff != b1.rrs[i].diff {
				return false
			}
			switch r0 := b0.rrs[i].r.(type) {
			case *bytesReader:
				switch r1 := b1.rrs[i].r.(type) {
				case *bytesReader:
					if !reflect.DeepEqual(r0.bs, r1.bs) || r0.index != r1.index {
						t.Logf("buffer differs: %+v, %+v", r0, r1)
						return false
					}
				default:
					t.Logf("buffer differs: %+v, %+v", r0, r1)
					return false
				}
			case *strings.Reader:
				switch r1 := b1.rrs[i].r.(type) {
				case *strings.Reader:
					if r0 != r1 {
						t.Logf("buffer differs: %+v, %+v", r0, r1)
						return false
					}
				default:
					t.Logf("buffer differs: %+v, %+v", r0, r1)
					return false
				}
			default:
				t.Logf("buffer differs: %+v, %+v", b0.rrs[i].r, b1.rrs[i].r)
				return false
			}
		}
		return true
	}

	if !bufferEqual(b1, b0) {
		t.Errorf("Buffer#Clone should be %+v but got %+v", b0, b1)
	}

	b1.Insert(4, 0x40)
	if bufferEqual(b1, b0) {
		t.Errorf("Buffer should not be equal: %+v, %+v", b0, b1)
	}

	b2 := b1.Clone()
	if !bufferEqual(b2, b1) {
		t.Errorf("Buffer#Clone should be %+v but got %+v", b1, b2)
	}

	b2.Replace(4, 0x40)
	b2.Flush()
	if !bufferEqual(b2, b1) {
		t.Errorf("Buffer should be equal: %+v, %+v", b1, b2)
	}

	b2.Replace(5, 0x40)
	b2.Flush()
	if bufferEqual(b2, b1) {
		t.Errorf("Buffer should not be equal: %+v, %+v", b1, b2)
	}
}

func TestBufferCopy(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))
	b.Replace(3, 0x41)
	b.Replace(4, 0x42)
	b.Replace(5, 0x43)
	b.Replace(9, 0x43)
	b.Replace(10, 0x44)
	b.Replace(11, 0x45)
	b.Replace(12, 0x46)
	b.Replace(14, 0x47)
	testCases := []struct {
		start, end int64
		expected   string
	}{
		{0, 16, "012ABC678CDEFdGf"},
		{0, 15, "012ABC678CDEFdG"},
		{1, 12, "12ABC678CDE"},
		{4, 14, "BC678CDEFd"},
		{2, 10, "2ABC678C"},
		{4, 10, "BC678C"},
		{2, 7, "2ABC6"},
		{5, 10, "C678C"},
		{7, 11, "78CD"},
		{8, 10, "8C"},
		{14, 20, "Gf"},
		{9, 9, ""},
		{10, 8, ""},
	}
	for _, testCase := range testCases {
		got := b.Copy(testCase.start, testCase.end)
		p := make([]byte, 17)
		_, _ = got.Read(p)
		if !strings.HasPrefix(string(p), testCase.expected+"\x00") {
			t.Errorf("Copy(%d, %d) should clone %q but got %q", testCase.start, testCase.end, testCase.expected, string(p))
		}
		got.Insert(0, 0x48)
		got.Insert(int64(len(testCase.expected)+1), 0x49)
		p = make([]byte, 19)
		_, _ = got.ReadAt(p, 0)
		if !strings.HasPrefix(string(p), "H"+testCase.expected+"I\x00") {
			t.Errorf("Copy(%d, %d) should clone %q but got %q", testCase.start, testCase.end, testCase.expected, string(p))
		}
	}
}

func TestBufferCut(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))
	b.Replace(3, 0x41)
	b.Replace(4, 0x42)
	b.Replace(5, 0x43)
	b.Replace(9, 0x43)
	b.Replace(10, 0x44)
	b.Replace(11, 0x45)
	b.Replace(12, 0x46)
	b.Replace(14, 0x47)
	testCases := []struct {
		start, end int64
		expected   string
	}{
		{0, 0, "012ABC678CDEFdGf"},
		{0, 4, "BC678CDEFdGf"},
		{0, 7, "78CDEFdGf"},
		{0, 10, "DEFdGf"},
		{0, 16, ""},
		{0, 20, ""},
		{3, 4, "012BC678CDEFdGf"},
		{3, 6, "012678CDEFdGf"},
		{3, 11, "012EFdGf"},
		{6, 10, "012ABCDEFdGf"},
		{6, 14, "012ABCGf"},
		{6, 15, "012ABCf"},
		{6, 17, "012ABC"},
		{8, 10, "012ABC67DEFdGf"},
		{8, 10, "012ABC67DEFdGf"},
		{10, 8, "012ABC678CDEFdGf"},
	}
	for _, testCase := range testCases {
		got := b.Clone()
		got.Cut(testCase.start, testCase.end)
		p := make([]byte, 17)
		_, _ = got.Read(p)
		if !strings.HasPrefix(string(p), testCase.expected+"\x00") {
			t.Errorf("Cut(%d, %d) should result into %q but got %q", testCase.start, testCase.end, testCase.expected, string(p))
		}
		got.Insert(0, 0x48)
		got.Insert(int64(len(testCase.expected)+1), 0x49)
		p = make([]byte, 19)
		_, _ = got.ReadAt(p, 0)
		if !strings.HasPrefix(string(p), "H"+testCase.expected+"I\x00") {
			t.Errorf("Cut(%d, %d) should result into %q but got %q", testCase.start, testCase.end, testCase.expected, string(p))
		}
	}
}

func TestBufferPaste(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))
	c := b.Copy(3, 13)
	b.Paste(5, c)
	p := make([]byte, 100)
	_, _ = b.ReadAt(p, 0)
	expected := "012343456789abc56789abcdef"
	if !strings.HasPrefix(string(p), expected+"\x00") {
		t.Errorf("p should be %q but got: %q", expected, string(p))
	}
	c.Replace(5, 0x41)
	c.Insert(6, 0x42)
	c.Insert(7, 0x43)
	b.Paste(10, c)
	p = make([]byte, 100)
	_, _ = b.ReadAt(p, 0)
	expected = "012343456734567ABC9abc89abc56789abcdef"
	if !strings.HasPrefix(string(p), expected+"\x00") {
		t.Errorf("p should be %q but got: %q", expected, string(p))
	}
	b.Cut(11, 14)
	b.Paste(13, c)
	b.Replace(13, 0x44)
	p = make([]byte, 100)
	_, _ = b.ReadAt(p, 0)
	expected = "012343456737AD4567ABC9abcBC9abc89abc56789abcdef"
	if !strings.HasPrefix(string(p), expected+"\x00") {
		t.Errorf("p should be %q but got: %q", expected, string(p))
	}
	b.Insert(14, 0x45)
	p = make([]byte, 100)
	_, _ = b.ReadAt(p, 0)
	expected = "012343456737ADE4567ABC9abcBC9abc89abc56789abcdef"
	if !strings.HasPrefix(string(p), expected+"\x00") {
		t.Errorf("p should be %q but got: %q", expected, string(p))
	}
}

func TestBufferInsert(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))

	tests := []struct {
		index    int64
		b        byte
		offset   int64
		expected string
		len      int64
	}{
		{0, 0x39, 0, "90123456", 17},
		{0, 0x38, 0, "89012345", 18},
		{4, 0x37, 0, "89017234", 19},
		{8, 0x30, 3, "17234056", 20},
		{9, 0x31, 3, "17234015", 21},
		{9, 0x32, 4, "72340215", 22},
		{23, 0x39, 19, "def9\x00\x00\x00\x00", 23},
		{23, 0x38, 19, "def89\x00\x00\x00", 24},
	}

	for _, test := range tests {
		b.Insert(test.index, test.b)
		p := make([]byte, 8)

		_, err := b.Seek(test.offset, io.SeekStart)
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}

		n, err := b.Read(p)
		if err != nil && err != io.EOF {
			t.Errorf("err should be nil or io.EOF but got: %v", err)
		}
		if n != len(strings.TrimRight(test.expected, "\x00")) {
			t.Errorf("n should be %d but got: %d", len(strings.TrimRight(test.expected, "\x00")), n)
		}
		if string(p) != test.expected {
			t.Errorf("p should be %s but got: %s", test.expected, string(p))
		}

		l, err := b.Len()
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if l != test.len {
			t.Errorf("l should be %d but got: %d", test.len, l)
		}
	}

	eis := b.EditedIndices()
	expected := []int64{0, 2, 4, 5, 8, 11, 23, 25}
	if !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if len(b.rrs) != 8 {
		t.Errorf("len(b.rrs) should be 8 but got: %d", len(b.rrs))
	}
}

func TestBufferReplace(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))

	tests := []struct {
		index    int64
		b        byte
		offset   int64
		expected string
		len      int64
	}{
		{0, 0x39, 0, "91234567", 16},
		{0, 0x38, 0, "81234567", 16},
		{1, 0x37, 0, "87234567", 16},
		{5, 0x30, 0, "87234067", 16},
		{4, 0x31, 0, "87231067", 16},
		{3, 0x30, 0, "87201067", 16},
		{2, 0x31, 0, "87101067", 16},
		{16, 0x31, 9, "9abcdef1", 17},
		{15, 0x30, 9, "9abcde01", 17},
		{2, 0x39, 0, "87901067", 17},
	}

	for _, test := range tests {
		b.Replace(test.index, test.b)
		p := make([]byte, 8)

		_, err := b.Seek(test.offset, io.SeekStart)
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}

		n, err := b.Read(p)
		if err != nil && err != io.EOF {
			t.Errorf("err should be nil or io.EOF but got: %v", err)
		}
		if n != 8 {
			t.Errorf("n should be 8 but got: %d", n)
		}
		if string(p) != test.expected {
			t.Errorf("p should be %s but got: %s", test.expected, string(p))
		}

		l, err := b.Len()
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if l != test.len {
			t.Errorf("l should be %d but got: %d", test.len, l)
		}
	}

	eis := b.EditedIndices()
	expected := []int64{0, 6, 15, 9223372036854775807}
	if !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if len(b.rrs) != 3 {
		t.Errorf("len(b.rrs) should be 4 but got: %d", len(b.rrs))
	}
}

func TestBufferDelete(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))

	tests := []struct {
		index    int64
		b        byte
		offset   int64
		expected string
		len      int64
	}{
		{4, 0x00, 0, "01235678", 15},
		{3, 0x00, 0, "01256789", 14},
		{6, 0x00, 0, "0125679a", 13},
		{0, 0x00, 0, "125679ab", 12},
		{4, 0x39, 0, "1256979a", 13},
		{5, 0x38, 0, "12569879", 14},
		{3, 0x00, 0, "1259879a", 13},
		{4, 0x00, 0, "125979ab", 12},
		{3, 0x00, 0, "12579abc", 11},
		{8, 0x39, 4, "9abc9def", 12},
		{8, 0x38, 4, "9abc89de", 13},
		{8, 0x00, 4, "9abc9def", 12},
		{8, 0x00, 4, "9abcdef\x00", 11},
	}

	for _, test := range tests {
		if test.b == 0x00 {
			b.Delete(test.index)
		} else {
			b.Insert(test.index, test.b)
		}
		p := make([]byte, 8)

		_, err := b.Seek(test.offset, io.SeekStart)
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}

		n, err := b.Read(p)
		if err != nil && err != io.EOF {
			t.Errorf("err should be nil or io.EOF but got: %v", err)
		}
		if n != len(strings.TrimRight(test.expected, "\x00")) {
			t.Errorf("n should be %d but got: %d", len(strings.TrimRight(test.expected, "\x00")), n)
		}
		if string(p) != test.expected {
			t.Errorf("p should be %s but got: %s", test.expected, string(p))
		}

		l, err := b.Len()
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if l != test.len {
			t.Errorf("l should be %d but got: %d", test.len, l)
		}
	}

	eis := b.EditedIndices()
	expected := []int64{}
	if !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if len(b.rrs) != 4 {
		t.Errorf("len(b.rrs) should be 4 but got: %d", len(b.rrs))
	}
}
