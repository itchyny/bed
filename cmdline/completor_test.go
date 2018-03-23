package cmdline

import (
	"testing"
)

func TestCompletorComplete(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := "new "
	cmd, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	if cmdline != "new Gopkg.toml" {
		t.Errorf("cmdline should be %q but got %q", "new Gopkg.toml", cmdline)
	}
	if c.target != "new " {
		t.Errorf("completion target should be %q but got %q", "new ", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	for i := 0; i < 3; i++ {
		cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	}
	if cmdline != "new .gitignore" {
		t.Errorf("cmdline should be %q but got %q", "new .gitignore", cmdline)
	}
	if c.target != "new " {
		t.Errorf("completion target should be %q but got %q", "new ", c.target)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	for i := 0; i < 4; i++ {
		cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	}
	if cmdline != "new editor/" {
		t.Errorf("cmdline should be %q but got %q", "new editor/", cmdline)
	}
	if c.target != "new " {
		t.Errorf("completion target should be %q but got %q", "new ", c.target)
	}
	if c.index != 7 {
		t.Errorf("completion index should be %d but got %d", 7, c.index)
	}

	cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	if cmdline != "new " {
		t.Errorf("cmdline should be %q but got %q", "new ", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	if cmdline != "new Gopkg.toml" {
		t.Errorf("cmdline should be %q but got %q", "new Gopkg.toml", cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(string(cmdline), cmd, prefix, arg, false)
	if cmdline != "new " {
		t.Errorf("cmdline should be %q but got %q", "new ", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for i := 0; i < 3; i++ {
		cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	}
	if cmdline != "new README.md" {
		t.Errorf("cmdline should be %q but got %q", "new README.md", cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 2, c.index)
	}
}

func TestCompletorCompleteKeepPrefix(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := " : : :  new   C"
	cmd, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	if cmdline != " : : :  new   cmdline/" {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   cmdline/", cmdline)
	}
	if c.target != " : : :  new   C" {
		t.Errorf("completion target should be %q but got %q", " : : :  new   C", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(string(cmdline), cmd, prefix, arg, true)
	if cmdline != " : : :  new   common/" {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   common/", cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(string(cmdline), cmd, prefix, arg, false)
	cmdline = c.complete(string(cmdline), cmd, prefix, arg, false)
	if cmdline != " : : :  new   C" {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   C", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}
}
