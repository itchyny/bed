package cmdline

import "os"

type env interface {
	Get(string) string
	List() []string
}

type environment struct{}

func (*environment) Get(key string) string {
	return os.Getenv(key)
}

func (*environment) List() []string {
	return os.Environ()
}
