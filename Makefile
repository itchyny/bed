BIN := bed
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -o $(BIN) ./cmd/$(BIN)

.PHONY: install
install:
	go install ./...

.PHONY: test
test: build
	go test -v -race -timeout 10s ./...

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

$(GOBIN)/golint:
	cd && go get golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf $(BIN)
	go clean ./...
