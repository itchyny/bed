package cmdline

import (
	"os"
	"time"
)

type mockFilesystem struct {
}

func (fs *mockFilesystem) Open(path string) (file, error) {
	return &mockFile{path}, nil
}

func (fs *mockFilesystem) Stat(path string) (os.FileInfo, error) {
	return &fileInfo{name: path, isDir: false}, nil
}

type mockFile struct {
	path string
}

func (f *mockFile) Close() error {
	return nil
}

func createFileInfoList(infos []*fileInfo) []os.FileInfo {
	fileInfos := make([]os.FileInfo, len(infos))
	for i, info := range infos {
		fileInfos[i] = info
	}
	return fileInfos
}

func (f *mockFile) Readdir(_ int) ([]os.FileInfo, error) {
	if f.path == "." {
		return createFileInfoList([]*fileInfo{
			{"README.md", false},
			{"Makefile", false},
			{".gitignore", false},
			{"Gopkg.toml", false},
			{"editor", true},
			{"cmdline", true},
			{"buffer", true},
			{"build", true},
		}), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	if f.path == homeDir {
		return createFileInfoList([]*fileInfo{
			{"Documents", true},
			{"Pictures", true},
			{"Library", true},
			{".vimrc", false},
			{".zshrc", false},
			{"example.txt", false},
		}), nil
	}
	if f.path == "/" {
		return createFileInfoList([]*fileInfo{
			{"bin", true},
			{"tmp", true},
			{"var", true},
			{"usr", true},
		}), nil
	}
	if f.path == "/bin" {
		return createFileInfoList([]*fileInfo{
			{"cp", false},
			{"echo", false},
			{"rm", false},
			{"ls", false},
			{"kill", false},
		}), nil
	}
	return nil, nil
}

type fileInfo struct {
	name  string
	isDir bool
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *fileInfo) Size() int64 {
	return 0
}

func (fi *fileInfo) Mode() os.FileMode {
	return os.FileMode(0x1ed)
}

func (fi *fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi *fileInfo) Sys() interface{} {
	return nil
}
