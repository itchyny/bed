package buffer

import (
	"io"
	"math"
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
	if expected := "01234567"; string(p) != expected {
		t.Errorf("p should be %q but got: %s", expected, string(p))
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
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if expected := "456789ab"; string(p) != expected {
		t.Errorf("p should be %q but got: %s", expected, string(p))
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
	if expected := "89abcdef"; string(p) != expected {
		t.Errorf("p should be %q but got: %s", expected, string(p))
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
	if expected := "cdefcdef"; string(p) != expected {
		t.Errorf("p should be %q but got: %s", expected, string(p))
	}

	n, err = b.ReadAt(p, 7)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if expected := "789abcde"; string(p) != expected {
		t.Errorf("p should be %q but got: %s", expected, string(p))
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
		for i := range len(b0.rrs) {
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
		if expected := len(strings.TrimRight(test.expected, "\x00")); n != expected {
			t.Errorf("n should be %d but got: %d", expected, n)
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
	if expected := []int64{0, 2, 4, 5, 8, 11, 23, 25}; !reflect.DeepEqual(eis, expected) {
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
		{15, 0x30, 8, "89abcde0", 16},
		{16, 0x31, 9, "9abcde01", 17},
		{2, 0x39, 0, "87901067", 17},
		{17, 0x32, 10, "abcde012", 18},
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
	if expected := []int64{0, 6, 15, math.MaxInt64}; !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if len(b.rrs) != 3 {
		t.Errorf("len(b.rrs) should be 3 but got: %d", len(b.rrs))
	}

	{
		b.Replace(3, 0x39)
		b.Replace(4, 0x38)
		b.Replace(5, 0x37)
		b.Replace(6, 0x36)
		b.Replace(7, 0x35)
		p := make([]byte, 8)
		if _, err := b.ReadAt(p, 2); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if expected := "99876589"; string(p) != expected {
			t.Errorf("p should be %s but got: %s", expected, string(p))
		}
		b.UndoReplace(7)
		b.UndoReplace(6)
		p = make([]byte, 8)
		if _, err := b.ReadAt(p, 2); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if expected := "99876789"; string(p) != expected {
			t.Errorf("p should be %s but got: %s", expected, string(p))
		}
		b.UndoReplace(5)
		b.UndoReplace(4)
		b.Flush()
		b.UndoReplace(3)
		b.UndoReplace(2)
		p = make([]byte, 8)
		if _, err := b.ReadAt(p, 2); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if expected := "99106789"; string(p) != expected {
			t.Errorf("p should be %s but got: %s", expected, string(p))
		}

		eis := b.EditedIndices()
		if expected := []int64{0, 6, 15, math.MaxInt64}; !reflect.DeepEqual(eis, expected) {
			t.Errorf("edited indices should be %v but got: %v", expected, eis)
		}
	}

	{
		b := NewBuffer(strings.NewReader("0123456789abcdef"))
		b.Replace(16, 0x30)
		b.Replace(10, 0x30)
		p := make([]byte, 8)
		if _, err := b.ReadAt(p, 9); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if expected := "90bcdef0"; string(p) != expected {
			t.Errorf("p should be %s but got: %s", expected, string(p))
		}

		l, _ := b.Len()
		if expected := int64(17); l != expected {
			t.Errorf("l should be %d but got: %d", expected, l)
		}

		eis := b.EditedIndices()
		if expected := []int64{10, 11, 16, math.MaxInt64}; !reflect.DeepEqual(eis, expected) {
			t.Errorf("edited indices should be %v but got: %v", expected, eis)
		}
	}
}

func TestBufferReplaceIn(t *testing.T) {
	b := NewBuffer(strings.NewReader("0123456789abcdef"))

	tests := []struct {
		start    int64
		end      int64
		b        byte
		offset   int64
		expected string
		len      int64
	}{
		{1, 2, 0x39, 0, "09234567", 16},
		{0, 6, 0x38, 0, "88888867", 16},
		{1, 3, 0x37, 0, "87788867", 16},
		{5, 7, 0x30, 0, "87788007", 16},
		{2, 6, 0x31, 0, "87111107", 16},
		{3, 4, 0x30, 0, "87101107", 16},
		{15, 16, 0x30, 8, "89abcde0", 16},
		{1, 5, 0x39, 0, "89999107", 16},
	}

	for _, test := range tests {
		b.ReplaceIn(test.start, test.end, test.b)
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
	if expected := []int64{0, 7, 15, 16}; !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if expected := 7; len(b.rrs) != expected {
		t.Errorf("len(b.rrs) should be %d but got: %d", expected, len(b.rrs))
	}

	{
		b := NewBuffer(strings.NewReader("0123456789abcdef"))
		b.ReplaceIn(16, 17, 0x30)
		b.ReplaceIn(10, 11, 0x30)
		p := make([]byte, 8)
		if _, err := b.ReadAt(p, 9); err != io.EOF {
			t.Errorf("err should be io.EOF but got: %v", err)
		}
		if expected := "90bcdef0"; string(p) != expected {
			t.Errorf("p should be %s but got: %s", expected, string(p))
		}

		l, _ := b.Len()
		if expected := int64(16); l != expected {
			t.Errorf("l should be %d but got: %d", expected, l)
		}

		eis := b.EditedIndices()
		if expected := []int64{10, 11, 16, 17}; !reflect.DeepEqual(eis, expected) {
			t.Errorf("edited indices should be %v but got: %v", expected, eis)
		}
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
		if expected := len(strings.TrimRight(test.expected, "\x00")); n != expected {
			t.Errorf("n should be %d but got: %d", expected, n)
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
	if expected := []int64{}; !reflect.DeepEqual(eis, expected) {
		t.Errorf("edited indices should be %v but got: %v", expected, eis)
	}

	if len(b.rrs) != 4 {
		t.Errorf("len(b.rrs) should be 4 but got: %d", len(b.rrs))
	}
}

func TestInsertInterval(t *testing.T) {
	tests := []struct {
		intervals   []int64
		newInterval []int64
		expected    []int64
	}{
		{[]int64{}, []int64{10, 20}, []int64{10, 20}},
		{[]int64{10, 20}, []int64{0, 5}, []int64{0, 5, 10, 20}},
		{[]int64{10, 20}, []int64{5, 15}, []int64{5, 20}},
		{[]int64{10, 20}, []int64{15, 17}, []int64{10, 20}},
		{[]int64{10, 20}, []int64{15, 25}, []int64{10, 25}},
		{[]int64{10, 20}, []int64{25, 30}, []int64{10, 20, 25, 30}},
		{[]int64{10, 20, 30, 40}, []int64{0, 5}, []int64{0, 5, 10, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 10}, []int64{5, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 15}, []int64{5, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 20}, []int64{5, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 25}, []int64{5, 25, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 30}, []int64{5, 40}},
		{[]int64{10, 20, 30, 40}, []int64{5, 45}, []int64{5, 45}},
		{[]int64{10, 20, 30, 40}, []int64{10, 20}, []int64{10, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{10, 30}, []int64{10, 40}},
		{[]int64{10, 20, 30, 40}, []int64{15, 45}, []int64{10, 45}},
		{[]int64{10, 20, 30, 40}, []int64{15, 25}, []int64{10, 25, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{15, 35}, []int64{10, 40}},
		{[]int64{10, 20, 30, 40}, []int64{35, 37}, []int64{10, 20, 30, 40}},
		{[]int64{10, 20, 30, 40}, []int64{40, 50}, []int64{10, 20, 30, 50}},
		{[]int64{10, 20, 30, 40, 50, 60, 70, 80}, []int64{45, 47}, []int64{10, 20, 30, 40, 45, 47, 50, 60, 70, 80}},
		{[]int64{10, 20, 30, 40, 50, 60, 70, 80}, []int64{35, 65}, []int64{10, 20, 30, 65, 70, 80}},
		{[]int64{10, 20, 30, 40, 50, 60, 70, 80}, []int64{25, 55}, []int64{10, 20, 25, 60, 70, 80}},
		{[]int64{10, 20, 30, 40, 50, 60, 70, 80}, []int64{75, 90}, []int64{10, 20, 30, 40, 50, 60, 70, 90}},
		{[]int64{10, 20, 30, 40, 50, 60, 70, 80}, []int64{0, 100}, []int64{0, 100}},
	}
	for _, test := range tests {
		got := insertInterval(test.intervals, test.newInterval[0], test.newInterval[1])
		if !reflect.DeepEqual(got, test.expected) {
			t.Errorf("insertInterval(%+v, %d, %d) should be %+v but got: %+v",
				test.intervals, test.newInterval[0], test.newInterval[1], test.expected, got)
		}
	}
}
