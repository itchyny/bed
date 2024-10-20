# Changelog
## [v0.2.7](https://github.com/itchyny/bed/compare/v0.2.6..v0.2.7) (2024-10-20)
* Support environment variable expansion in the command line.
* Implement `:cd`, `:chdir`, `:pwd` commands to change the working directory.
* Improve command line completion for command name and environment variables.
* Recognize file name argument and bang for `:wq` command.

## [v0.2.6](https://github.com/itchyny/bed/compare/v0.2.5..v0.2.6) (2024-10-08)
* Support reading from standard input.
* Implement command line history.

## [v0.2.5](https://github.com/itchyny/bed/compare/v0.2.4..v0.2.5) (2024-05-03)
* Require Go 1.22.

## [v0.2.4](https://github.com/itchyny/bed/compare/v0.2.3..v0.2.4) (2023-09-30)
* Require Go 1.21.

## [v0.2.3](https://github.com/itchyny/bed/compare/v0.2.2..v0.2.3) (2022-12-25)
* Fix crash on window moving commands on the last window.

## [v0.2.2](https://github.com/itchyny/bed/compare/v0.2.1..v0.2.2) (2021-09-14)
* Add `:only` command to make the current window the only one.
* Reduce memory allocations on rendering.
* Release `arm64` artifacts.

## [v0.2.1](https://github.com/itchyny/bed/compare/v0.2.0..v0.2.1) (2020-12-29)
* Add `:{count}%` to go to the position by percentage in the file.
* Add `:{count}go[to]` command to go to the specific line.

## [v0.2.0](https://github.com/itchyny/bed/compare/v0.1.0..v0.2.0) (2020-04-10)
* Add `:cquit` command.

## [v0.1.0](https://github.com/itchyny/bed/compare/8239ec4..v0.1.0) (2020-01-25)
* Initial implementation.
