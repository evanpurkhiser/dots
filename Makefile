.PHONY: dist/dots

VERSION := $(shell git describe)

dist/dots:
	rm -f dist/dots
	go build -ldflags "-X main.Version=$(VERSION)" -o dist/dots cmd/dots/*.go
