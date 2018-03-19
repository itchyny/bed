package main

import (
	"fmt"
	"os"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/editor"
	"github.com/itchyny/bed/tui"
	"github.com/itchyny/bed/window"
)

func run(args []string) int {
	if len(args) > 2 {
		fmt.Fprintf(os.Stderr, "%s: too many files\n", name)
		return 1
	}
	editor := editor.NewEditor(
		tui.NewTui(), window.NewManager(), cmdline.NewCmdline(),
	)
	if err := editor.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return 1
	}
	if len(args) > 1 {
		if err := editor.Open(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
			return 1
		}
	} else {
		if err := editor.OpenEmpty(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
			return 1
		}
	}
	if err := editor.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return 1
	}
	if err := editor.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return 1
	}
	return 0
}
