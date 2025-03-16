BINARY = godnsvalidator
VERSION = 0.1.0

.PHONY: build
build:
	go build -o bin/$(BINARY) ./cmd/godnsvalidator

.PHONY: install
install:
	go install ./cmd/godnsvalidator

.PHONY: release
release:
	goreleaser release --rm-dist

.PHONY: clean
clean:
	rm -rf bin/ dist/