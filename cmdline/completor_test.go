package cmdline

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompletorCompleteFilepath(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := "new "
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
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
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
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
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "new editor"+string(filepath.Separator) {
		t.Errorf("cmdline should be %q but got %q", "new editor"+string(filepath.Separator), cmdline)
	}
	if c.target != "new " {
		t.Errorf("completion target should be %q but got %q", "new ", c.target)
	}
	if c.index != 7 {
		t.Errorf("completion index should be %d but got %d", 7, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "new " {
		t.Errorf("cmdline should be %q but got %q", "new ", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "new Gopkg.toml" {
		t.Errorf("cmdline should be %q but got %q", "new Gopkg.toml", cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if cmdline != "new " {
		t.Errorf("cmdline should be %q but got %q", "new ", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for i := 0; i < 3; i++ {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "new README.md" {
		t.Errorf("cmdline should be %q but got %q", "new README.md", cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 2, c.index)
	}

	c.clear()
	cmdline = "new Gopkg.to"
	cmd, _, prefix, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "new Gopkg.toml" {
		t.Errorf("cmdline should be %q but got %q", "new Gopkg.toml", cmdline)
	}
	if c.target != "new Gopkg.to" {
		t.Errorf("completion target should be %q but got %q", "new Gopkg.to", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = "edit"
	cmd, _, prefix, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "edit Gopkg.toml" {
		t.Errorf("cmdline should be %q but got %q", "edit Gopkg.toml", cmdline)
	}
	if c.target != "edit" {
		t.Errorf("completion target should be %q but got %q", "edit", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathKeepPrefix(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := " : : :  new   B"
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != " : : :  new   buffer"+string(filepath.Separator) {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   buffer"+string(filepath.Separator), cmdline)
	}
	if c.target != " : : :  new   B" {
		t.Errorf("completion target should be %q but got %q", " : : :  new   B", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != " : : :  new   build"+string(filepath.Separator) {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   build"+string(filepath.Separator), cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if cmdline != " : : :  new   B" {
		t.Errorf("cmdline should be %q but got %q", " : : :  new   B", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}
}

func TestCompletorCompleteFilepathHomedir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{})
	cmdline := "vnew ~/"
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "vnew ~/example.txt" {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/example.txt", cmdline)
	}
	if c.target != "vnew ~/" {
		t.Errorf("completion target should be %q but got %q", "vnew ~/", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "vnew ~/.vimrc" {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/.vimrc", cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	for i := 0; i < 3; i++ {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "vnew ~/Library/" {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/Library/", cmdline)
	}
	if c.index != 4 {
		t.Errorf("completion index should be %d but got %d", 4, c.index)
	}

	for i := 0; i < 2; i++ {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "vnew ~/" {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}
}

func TestCompletorCompleteFilepathHomedirDot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{})
	cmdline := "vnew ~/."
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if cmdline != "vnew ~/.zshrc" {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/.zshrc", cmdline)
	}
	if c.target != "vnew ~/." {
		t.Errorf("completion target should be %q but got %q", "vnew ~/.", c.target)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "vnew ~/." {
		t.Errorf("cmdline should be %q but got %q", "vnew ~/.", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}
}

func TestCompletorCompleteFilepathRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{})
	cmdline := "e /"
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "e /bin/" {
		t.Errorf("cmdline should be %q but got %q", "e /bin/", cmdline)
	}
	if c.target != "e /" {
		t.Errorf("completion target should be %q but got %q", "e /", c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "e /tmp/" {
		t.Errorf("cmdline should be %q but got %q", "e /tmp/", cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	c.clear()
	cmd, _, prefix, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "e /bin/cp" {
		t.Errorf("cmdline should be %q but got %q", "e /bin/cp", cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteWincmd(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := "winc"
	cmd, _, prefix, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "winc" {
		t.Errorf("cmdline should be %q but got %q", "winc", cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for i := 0; i < 4; i++ {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "winc k" {
		t.Errorf("cmdline should be %q but got %q", "winc k", cmdline)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	for i := 0; i < 5; i++ {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if cmdline != "winc J" {
		t.Errorf("cmdline should be %q but got %q", "winc J", cmdline)
	}
	if c.index != 8 {
		t.Errorf("completion index should be %d but got %d", 8, c.index)
	}

	c.clear()
	cmd, _, prefix, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if cmdline != "winc J" {
		t.Errorf("cmdline should be %q but got %q", "winc J", cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}
