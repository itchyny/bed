package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"

	"golang.org/x/term"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/editor"
	"github.com/itchyny/bed/tui"
	"github.com/itchyny/bed/window"
)

const name = "bed"

const version = "0.2.8"

var revision = "HEAD"

func main() {
	os.Exit(run(os.Args[1:]))
}

const (
	exitCodeOK = iota
	exitCodeErr
)

func run(args []string) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fs.SetOutput(os.Stdout)
		fmt.Printf(`%[1]s - binary editor written in Go

Version: %s (rev: %s/%s)

Synopsis:
  %% %[1]s file

Options:
`, name, version, revision, runtime.Version())
		fs.PrintDefaults()
	}
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "print version")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return exitCodeOK
		}
		return exitCodeErr
	}
	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return exitCodeOK
	}
	if err := start(fs.Args()); err != nil {
		if err, ok := err.(interface{ ExitCode() int }); ok {
			return err.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return exitCodeErr
	}
	return exitCodeOK
}

func start(args []string) error {
	if len(args) > 1 {
		return errors.New("too many files")
	}
	editor := editor.NewEditor(
		tui.NewTui(), window.NewManager(), cmdline.NewCmdline(),
	)
	if err := editor.Init(); err != nil {
		return err
	}
	if len(args) > 0 && args[0] != "-" {
		if err := editor.Open(args[0]); err != nil {
			return err
		}
	} else if term.IsTerminal(int(os.Stdin.Fd())) {
		if err := editor.OpenEmpty(); err != nil {
			return err
		}
	} else {
		if err := editor.Read(os.Stdin); err != nil {
			return err
		}
	}
	defer editor.Close()
	return editor.Run()
}
