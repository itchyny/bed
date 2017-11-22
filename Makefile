BIN = bed

all: clean lint build test

build: deps
	go build -o build/$(BIN) ./cmd/...

install: deps
	go install ./...

deps:
	command -v dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure

test: build
	go test -v ./...

lint: lintdeps build
	golint -set_exit_status $$(go list ./... | grep -v /vendor/)

lintdeps:
	go get -d -v -t ./...
	go get -u github.com/golang/lint/golint

clean:
	rm -rf build
	go clean

.PHONY: build install deps test
