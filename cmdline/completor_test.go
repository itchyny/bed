package cmdline

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestCompletorCompleteCommand(t *testing.T) {
	c := newCompletor(nil, nil)
	cmdline := c.complete("", true)
	if expected := "cd"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
	if expected := "edit"; !slices.Contains(c.results, expected) {
		t.Errorf("completion results should contain %q but got %v", expected, c.results)
	}
	if expected := "goto"; slices.Contains(c.results, expected) {
		t.Errorf("completion results should not contain %q but got %v", expected, c.results)
	}
	if expected := "write"; !slices.Contains(c.results, expected) {
		t.Errorf("completion results should contain %q but got %v", expected, c.results)
	}

	for range 3 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "edit"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}

	for range 4 {
		cmdline = c.complete(cmdline, false)
	}
	if expected := ""; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	for range 3 {
		cmdline = c.complete(cmdline, false)
	}
	if expected := "write"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}

	c.clear()
	cmdline = c.complete(": :\t", true)
	if expected := ": :\tcd"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}

	c.clear()
	cmdline = c.complete(": : cq", true)
	if expected := ": : cquit"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}

	c.clear()
	cmdline = c.complete("e", false)
	if expected := "exit"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	cmdline = c.complete(cmdline, true)
	if expected := "e"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	cmdline = c.complete(cmdline, true)
	if expected := "edit"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	cmdline = c.complete(cmdline, false)
	if expected := "e"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}

	c.clear()
	cmdline = "p"
	for _, expected := range []string{"pwd", "pwd"} {
		cmdline = c.complete(cmdline, true)
		if cmdline != expected {
			t.Errorf("cmdline should be %q but got %q", expected, cmdline)
		}
	}

	c.clear()
	cmdline = "10"
	for _, command := range []string{"%", "goto", ""} {
		cmdline = c.complete(cmdline, true)
		if expected := "10" + command; cmdline != expected {
			t.Errorf("cmdline should be %q but got %q", expected, cmdline)
		}
	}

	c.clear()
	cmdline = "10,20"
	for _, command := range []string{"wq", "write", "xall", "xit", ""} {
		cmdline = c.complete(cmdline, true)
		if expected := "10,20" + command; cmdline != expected {
			t.Errorf("cmdline should be %q but got %q", expected, cmdline)
		}
	}

	c.clear()
	cmdline = c.complete("not", true)
	if expected := "not"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if len(c.results) != 0 {
		t.Errorf("completion results should be empty but got %v", c.results)
	}
}

func TestCompletorCompleteFilepath(t *testing.T) {
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("new", true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if expected := "README.md"; !slices.Contains(c.results, expected) {
		t.Errorf("completion results should contain %q but got %v", expected, c.results)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "new .gitignore"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 3 {
		t.Errorf("completion index should be %d but got %d", 3, c.index)
	}

	for range 4 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "new editor" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 7 {
		t.Errorf("completion index should be %d but got %d", 7, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := "new"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, false)
	if expected := "new"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "new README.md"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 2, c.index)
	}

	c.clear()
	cmdline = c.complete("new Gopkg.to", true)
	if expected := "new Gopkg.toml"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := ""; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("new not", true)
	if expected := "new not"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "new not"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("edit", true)
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
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("edit .", true)
	if expected := "edit .gitignore"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := ""; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("edit ./r", true)
	if expected := "edit ./README.md"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := ""; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("cd ..", true)
	if expected := "cd ../"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := ""; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathKeepPrefix(t *testing.T) {
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete(" : : :  new   \tB", true)
	if expected := " : : :  new   \tbuffer" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := " : : :  new   \tB"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := " : : :  new   \tbuild" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	for range 2 {
		cmdline = c.complete(cmdline, false)
	}
	if expected := " : : :  new   \tB"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	c.clear()
	cmdline = c.complete(" : cd\u3000", true)
	if expected := " : cd\u3000buffer" + string(filepath.Separator); cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathHomedir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("vnew ~/", true)
	if expected := "vnew ~/example.txt"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "vnew ~/"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := "vnew ~/.vimrc"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	for range 3 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "vnew ~/Library/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 4 {
		t.Errorf("completion index should be %d but got %d", 4, c.index)
	}

	for range 2 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "vnew ~/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	c.clear()
	cmdline = c.complete("cd ~user/", true)
	if expected := "cd ~user/Documents/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "cd ~user/"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathHomedirDot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("vnew ~/.", false)
	if expected := "vnew ~/.zshrc"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "vnew ~/."; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := "vnew ~/."; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}
}

func TestCompletorCompleteFilepathEnviron(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}

	c := newCompletor(&mockFilesystem{}, &mockEnvironment{})
	cmdline := c.complete("e $h", true)
	if expected := "e $HOME/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("e $HOME/", true)
	if expected := "e $HOME/example.txt"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "e $HOME/"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("cd $h", true)
	if expected := "cd $HOME/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	c.clear()
	cmdline = c.complete("cd $HOME/", true)
	if expected := "cd $HOME/Documents/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteFilepathRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("e /", true)
	if expected := "e /bin/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if expected := "e /"; c.target != expected {
		t.Errorf("completion target should be %q but got %q", expected, c.target)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, true)
	if expected := "e /tmp/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 1 {
		t.Errorf("completion index should be %d but got %d", 1, c.index)
	}

	cmdline = c.complete(cmdline, false)
	c.clear()
	cmdline = c.complete(cmdline, true)
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
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("cd ", false)
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
	cmdline = c.complete("cd ~/", false)
	if expected := "cd ~/Pictures/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 2 {
		t.Errorf("completion index should be %d but got %d", 2, c.index)
	}

	c.clear()
	cmdline = c.complete("cd /", true)
	if expected := "cd /bin/"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}

func TestCompletorCompleteWincmd(t *testing.T) {
	c := newCompletor(&mockFilesystem{}, nil)
	cmdline := c.complete("winc", true)
	if expected := "wincmd n"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	for range 7 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "wincmd b"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 7 {
		t.Errorf("completion index should be %d but got %d", 7, c.index)
	}

	for range 7 {
		cmdline = c.complete(cmdline, true)
	}
	if expected := "wincmd n"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}

	cmdline = c.complete(cmdline, false)
	if expected := "wincmd"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != -1 {
		t.Errorf("completion index should be %d but got %d", -1, c.index)
	}

	c.clear()
	cmdline = c.complete("winc j", true)
	if expected := "winc j"; cmdline != expected {
		t.Errorf("cmdline should be %q but got %q", expected, cmdline)
	}
	if c.index != 0 {
		t.Errorf("completion index should be %d but got %d", 0, c.index)
	}
}
