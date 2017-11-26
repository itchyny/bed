package core

import "testing"

func TestKeyManagerPress(t *testing.T) {
	km := NewKeyManager()
	km.Register(CursorUp, "k")
	e := km.Press("k")
	if e.Type != CursorUp {
		t.Errorf("pressing k should emit CursorUp but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != Nop {
		t.Errorf("pressing j should be nop but got: %d", e.Type)
	}
}

func TestKeyManagerPressMulti(t *testing.T) {
	km := NewKeyManager()
	km.Register(CursorUp, "k", "k", "j")
	km.Register(CursorDown, "k", "j", "j")
	km.Register(CursorDown, "j", "k", "k")
	e := km.Press("k")
	if e.Type != Nop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != Nop {
		t.Errorf("pressing k twice should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != Nop {
		t.Errorf("pressing k three times should be nop but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != CursorUp {
		t.Errorf("pressing kkj should emit CursorUp but got: %d", e.Type)
	}
}

func TestKeyManagerPressCount(t *testing.T) {
	km := NewKeyManager()
	km.Register(CursorUp, "k", "j")
	e := km.Press("k")
	if e.Type != Nop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("3")
	if e.Type != Nop {
		t.Errorf("pressing 3 should be nop but got: %d", e.Type)
	}
	e = km.Press("7")
	if e.Type != Nop {
		t.Errorf("pressing 7 should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != Nop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != CursorUp {
		t.Errorf("pressing 37kj should emit CursorUp but got: %d", e.Type)
	}
	if e.Count != 37 {
		t.Errorf("pressing 37kj should emit CursorUp with count 37 but got: %d", e.Count)
	}
}
