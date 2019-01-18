package main

import (
	"fmt"
	"os"
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
