BIN = bed

.PHONY: all
all: clean build

.PHONY: build
build: deps
	go build -o build/$(BIN) ./cmd/...

.PHONY: install
install: deps
	go install ./...

.PHONY: deps
deps:
	command -v dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure

.PHONY: test
test: build
	@! git grep tcell -- ':!tui/' ':!Gopkg.lock' ':!Gopkg.toml' ':!Makefile'
	go test -v ./...

.PHONY: lint
lint: lintdeps build
	golint -set_exit_status $$(go list ./... | grep -v /vendor/)

.PHONY: lintdeps
lintdeps:
	go get -d -v -t ./...
	command -v golint >/dev/null || go get -u golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf build
	rm -rf vendor
	go clean
