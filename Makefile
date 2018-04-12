BIN = bed

all: clean build

build: deps
	go build -o build/$(BIN) ./cmd/...

install: deps
	go install ./...

deps:
	command -v dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure
	go get -u github.com/gdamore/tcell

test: build
	@! git grep tcell -- ':!tui/' ':!Gopkg.lock' ':!Gopkg.toml' ':!Makefile'
	go test -v ./...

lint: lintdeps build
	golint -set_exit_status $$(go list ./... | grep -v /vendor/)

lintdeps:
	go get -d -v -t ./...
	command -v golint >/dev/null || go get -u github.com/golang/lint/golint

clean:
	rm -rf build
	go clean

.PHONY: build install deps test lint lintdeps clean
