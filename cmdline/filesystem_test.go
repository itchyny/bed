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
			&fileInfo{"README.md", false},
			&fileInfo{"Makefile", false},
			&fileInfo{".gitignore", false},
			&fileInfo{"Gopkg.toml", false},
			&fileInfo{"editor", true},
			&fileInfo{"cmdline", true},
			&fileInfo{"buffer", true},
			&fileInfo{"common", true},
		}), nil
	}
	if f.path == "/" {
		return nil, nil
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
