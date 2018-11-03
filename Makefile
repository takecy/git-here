.PHONY: prepare build update release test

prepare:
	brew install goreleaser/tap/goreleaser

build:
	GO111MODULE=on go build -o gh_dev ./gh

vendor: update
	GO111MODULE=on go mod vendor

update:
	GO111MODULE=on go get -u

release:
	GO111MODULE=on GITHUB_TOKEN=$${GITHUB_TOKEN} goreleaser --rm-dist

test:
	GO111MODULE=on go test -v -race ./...