# bed
[![CI Status](https://github.com/itchyny/bed/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/itchyny/bed/actions?query=branch:main)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchyny/bed)](https://goreportcard.com/report/github.com/itchyny/bed)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchyny/bed/blob/main/LICENSE)
[![release](https://img.shields.io/github/release/itchyny/bed/all.svg)](https://github.com/itchyny/bed/releases)
[![pkg.go.dev](https://pkg.go.dev/badge/github.com/itchyny/bed)](https://pkg.go.dev/github.com/itchyny/bed)

Binary editor written in Go

## Screenshot
![bed command screenshot](https://user-images.githubusercontent.com/375258/38499347-2f71306c-3c42-11e8-926e-1782b0bc73f3.png)

## Motivation
I wanted to create a binary editor with Vim-like user interface, which runs in terminals, fast, and is portable.
I have always been interested in various binary formats and I wanted to create my own editor to handle them.
I also wanted to learn how a binary editor can handle large files and allow users to edit them interactively.

While creating this binary editor, I leaned a lot about programming in Go language.
I spent a lot of time writing the core logic of buffer implementation of the editor.
It was a great learning experience for me and a lot of fun.

## Installation
### Homebrew

```sh
brew install bed
```

### Build from source

```bash
go install github.com/itchyny/bed/cmd/bed@latest
```

## Features

- Basic byte editing
- Large file support
- Command line interface
- Window splitting
- Partial writing
- Text searching
- Undo and redo

### Commands and keyboard shortcuts
This binary editor is influenced by the Vim editor.

- File operations
  - `:edit`, `:enew`, `:new`, `:vnew`, `:only`
- Current working directory
  - `:cd`, `:chdir`, `:pwd`
- Quit and save
  - `:quit`, `ZQ`, `:qall`, `:write`,
    `:wq`, `ZZ`, `:xit`, `:xall`, `:cquit`
- Window operations
  - `:wincmd [nohjkltbpHJKL]`, `<C-w>[nohjkltbpHJKL]`
- Cursor motions
  - `h`, `j`, `k`, `l`, `w`, `b`, `^`, `0`, `$`,
    `<C-[fb]>`, `<C-[du]>`, `<C-[ey]>`, `<C-[np]>`,
    `G`, `gg`, `:{count}`, `:{count}goto`, `:{count}%`,
    `H`, `M`, `L`, `zt`, `zz`, `z.`, `zb`, `z-`,
    `<TAB>` (toggle focus between hex and text views)
- Mode operations
  - `i`, `I`, `a`, `A`, `v`, `r`, `R`, `<ESC>`
- Inspect and edit
  - `gb` (binary), `gd` (decimal), `x` (delete), `X` (delete backward),
    `d` (delete selection), `y` (copy selection), `p`, `P` (paste),
    `<` (left shift), `>` (right shift), `<C-a>` (increment), `<C-x>` (decrement)
- Undo and redo
  - `:undo`, `u`, `:redo`, `<C-r>`
- Search
  - `/`, `?`, `n`, `N`, `<C-c>` (abort)

## Bug Tracker
Report bug at [Issues・itchyny/bed - GitHub](https://github.com/itchyny/bed/issues).

## Author
itchyny (<https://github.com/itchyny>)

## License
This software is released under the MIT License, see LICENSE.
