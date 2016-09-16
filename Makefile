build:
	go get github.com/mitchellh/gox
	gox -os="darwin linux windows"

.PHONY: build
