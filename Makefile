BIN := bed
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -o build/$(BIN) ./cmd/$(BIN)

.PHONY: install
install:
	go install ./...

.PHONY: test
test: build
	@! git grep tcell -- ':!tui/' ':!Makefile' ':!go.sum' ':!go.mod'
	go test -v ./...

.PHONY: lint
lint: lintdeps
	go vet ./...
	golint -set_exit_status ./...

.PHONY: lintdeps
lintdeps:
	GO111MODULE=off go get golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf build
	go clean ./...
