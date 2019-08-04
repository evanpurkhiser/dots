.PHONY: dist/dots

dist/dots:
	rm -f dist/dots
	go build -o dist/dots cmd/dots/*.go
