package cmdline

import "os"

type fs interface {
	Open(string) (file, error)
	Stat(string) (os.FileInfo, error)
}

type file interface {
	Close() error
	Readdir(int) ([]os.FileInfo, error)
}

type filesystem struct {
}

func (fs *filesystem) Open(path string) (file, error) {
	return os.Open(path)
}

func (fs *filesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
