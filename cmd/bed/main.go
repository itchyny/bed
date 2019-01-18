package main

import (
	"fmt"
	"os"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/editor"
	"github.com/itchyny/bed/tui"
	"github.com/itchyny/bed/window"
)

const cmdName = "bed"
const version = "v0.0.0"
const author = "itchyny"

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("too many files")
	}
	editor := editor.NewEditor(
		tui.NewTui(), window.NewManager(), cmdline.NewCmdline(),
	)
	if err := editor.Init(); err != nil {
		return err
	}
	if len(args) > 1 {
		if err := editor.Open(args[1]); err != nil {
			return err
		}
	} else {
		if err := editor.OpenEmpty(); err != nil {
			return err
		}
	}
	if err := editor.Run(); err != nil {
		return err
	}
	if err := editor.Close(); err != nil {
		return err
	}
	return nil
}
