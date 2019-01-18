package main

import (
	"fmt"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/editor"
	"github.com/itchyny/bed/tui"
	"github.com/itchyny/bed/window"
)

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
