package main

import (
	"fmt"
	"os"

	"github.com/itchyny/bed/core"
)

func run(args []string) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "%s: specify a filename\n", name)
		return 1
	}
	if len(args) > 2 {
		fmt.Fprintf(os.Stderr, "%s: too many files\n", name)
		return 1
	}
	editor := core.NewEditor()
	if err := editor.Open(args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		return 1
	}
	return 0
}
