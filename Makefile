BIN := bed
export GO111MODULE=on

.PHONY: all
all: clean build

.PHONY: build
build: deps
	go build -o build/$(BIN) ./cmd/$(BIN)

.PHONY: install
install: deps
	go install ./...

.PHONY: deps
deps:
	go get -d -v ./...

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
	GO111MODULE=off go get -u golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf build
	go clean ./...
