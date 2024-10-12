package cmdline

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompletorCompleteFilepath(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := "new "
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new "; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "new .gitignore"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new "; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	for range 4 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "new editor" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new "; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 7 {
		t.Errorf("completion index should be %d but got %d", 7, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "new "; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if expected := "new "; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "new README.md"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 2, c.index)
	}

	c.clear()
	cmdline = "new Gopkg.to"
	cmd, _, prefix, _, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new Gopkg.to"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = "edit"
	cmd, _, prefix, _, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "edit Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "edit"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathLeadingDot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{})
	cmdline := "edit ."
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "edit .gitignore"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "edit ."; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = "edit ./r"
	cmd, _, prefix, _, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "edit ./README.md"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "edit ./r"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathKeepPrefix(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := " : : :  new   B"
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := " : : :  new   buffer" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := " : : :  new   B"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := " : : :  new   build" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if expected := " : : :  new   B"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
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
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "vnew ~/example.txt"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "vnew ~/"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "vnew ~/.vimrc"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "vnew ~/Library/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 4 {
		t.Errorf("completion index should be %d but got %d", 4, c.index)
	}

	for range 2 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "vnew ~/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
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
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if expected := "vnew ~/.zshrc"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "vnew ~/."; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "vnew ~/."; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
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
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "e /bin/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "e /"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "e /tmp/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	c.clear()
	cmd, _, prefix, _, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "e /bin/cp"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathChdir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{})
	cmdline := "cd "
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, false)
	if expected := "cd editor/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "cd "; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	c.clear()
	cmdline = c.complete(cmdline, cmd, prefix, "~/", false)
	if expected := "cd ~/Pictures/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete(cmdline, cmd, prefix, "/", true)
	if expected := "cd /bin/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteWincmd(t *testing.T) {
	c := newCompletor(&mockFilesystem{})
	cmdline := "winc"
	cmd, _, prefix, _, arg, _ := parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "winc"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for range 4 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "winc k"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	for range 5 {
		cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	}
	if expected := "winc J"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 8 {
		t.Errorf("completion index should be %d but got %d", 8, c.index)
	}

	c.clear()
	cmd, _, prefix, _, arg, _ = parse([]rune(cmdline))
	cmdline = c.complete(cmdline, cmd, prefix, arg, true)
	if expected := "winc J"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}
