build:
	go build -o gs_dev ./gs

build_c:
	go get -u github.com/mitchellh/gox
	gox ./gs

update:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v -update

test:
	go test -race ./...