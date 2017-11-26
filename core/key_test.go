package core

import "testing"

func TestKeyManagerPress(t *testing.T) {
	km := NewKeyManager()
	km.Register(CursorUp, "k")
	e := km.Press("k")
	if e != CursorUp {
		t.Errorf("pressing k should emit CursorUp but got: %d", e)
	}
	e = km.Press("j")
	if e != Nop {
		t.Errorf("pressing j should be nop but got: %d", e)
	}
}

func TestKeyManagerPressMulti(t *testing.T) {
	km := NewKeyManager()
	km.Register(CursorUp, "k", "k", "j")
	km.Register(CursorDown, "k", "j", "j")
	km.Register(CursorDown, "j", "k", "k")
	e := km.Press("k")
	if e != Nop {
		t.Errorf("pressing k should be nop but got: %d", e)
	}
	e = km.Press("k")
	if e != Nop {
		t.Errorf("pressing k twice should be nop but got: %d", e)
	}
	e = km.Press("k")
	if e != Nop {
		t.Errorf("pressing k three times should be nop but got: %d", e)
	}
	e = km.Press("j")
	if e != CursorUp {
		t.Errorf("pressing kkj should emit CursorUp but got: %d", e)
	}
}
