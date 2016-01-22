build:
	go build .

build_c:
	go get github.com/mitchellh/gox
	gox ./...
