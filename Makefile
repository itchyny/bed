BIN = bed

all: clean test build

build: deps
	go build -o build/$(BIN) ./cmd/...

install: deps
	go install ./...

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

test: build
	go test -v ./...

clean:
	rm -rf build
	go clean

.PHONY: build install deps test
