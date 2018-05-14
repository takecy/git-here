.PHONY: prepare build build_c update test

prepare:
	go get -u github.com/golang/dep/cmd/dep

build:
	go build -o gs_dev ./gs

build_c:
	go get -u github.com/mitchellh/gox
	gox ./gs

update:
	dep ensure -v -update

release:
	GITHUB_TOKEN=$${GITHUB_TOKEN} goreleaser --rm-dist

test:
	go test -race ./...