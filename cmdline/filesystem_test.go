package cmdline

import (
	"os"
	"os/user"
	"time"
)

const mockHomeDir = "/home/user"

type mockFilesystem struct{}

func (*mockFilesystem) Open(path string) (file, error) {
	return &mockFile{path}, nil
}

func (*mockFilesystem) Stat(path string) (os.FileInfo, error) {
	return &mockFileInfo{name: path, isDir: path == mockHomeDir}, nil
}

func (*mockFilesystem) GetUser(name string) (*user.User, error) {
	return &user.User{Username: name, HomeDir: mockHomeDir}, nil
}

func (*mockFilesystem) UserHomeDir() (string, error) {
	return mockHomeDir, nil
}

type mockFile struct {
	path string
}

func (*mockFile) Close() error {
	return nil
}

func createFileInfoList(infos []*mockFileInfo) []os.FileInfo {
	fileInfos := make([]os.FileInfo, len(infos))
	for i, info := range infos {
		fileInfos[i] = info
	}
	return fileInfos
}

func (f *mockFile) Readdir(_ int) ([]os.FileInfo, error) {
	if f.path == "." {
		return createFileInfoList([]*mockFileInfo{
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
	if f.path == mockHomeDir {
		return createFileInfoList([]*mockFileInfo{
			{"Documents", true},
			{"Pictures", true},
			{"Library", true},
			{".vimrc", false},
			{".zshrc", false},
			{"example.txt", false},
		}), nil
	}
	if f.path == "/" {
		return createFileInfoList([]*mockFileInfo{
			{"bin", true},
			{"tmp", true},
			{"var", true},
			{"usr", true},
		}), nil
	}
	if f.path == "/bin" {
		return createFileInfoList([]*mockFileInfo{
			{"cp", false},
			{"echo", false},
			{"rm", false},
			{"ls", false},
			{"kill", false},
		}), nil
	}
	return nil, nil
}

type mockFileInfo struct {
	name  string
	isDir bool
}

func (fi *mockFileInfo) Name() string {
	return fi.name
}

func (fi *mockFileInfo) IsDir() bool {
	return fi.isDir
}

func (*mockFileInfo) Size() int64 {
	return 0
}

func (*mockFileInfo) Mode() os.FileMode {
	return os.FileMode(0x1ed)
}

func (*mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (*mockFileInfo) Sys() any {
	return nil
}
