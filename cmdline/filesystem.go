package cmdline

import (
	"os"
	"os/user"
)

type fs interface {
	Open(string) (file, error)
	Stat(string) (os.FileInfo, error)
	GetUser(string) (*user.User, error)
	UserHomeDir() (string, error)
}

type file interface {
	Close() error
	Readdir(int) ([]os.FileInfo, error)
}

type filesystem struct{}

func (*filesystem) Open(path string) (file, error) {
	return os.Open(path)
}

func (*filesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (*filesystem) GetUser(name string) (*user.User, error) {
	return user.Lookup(name)
}

func (*filesystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
