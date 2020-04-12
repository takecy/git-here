.PHONY: prepare build update release test

prepare:
	brew install goreleaser/tap/goreleaser

build:
	GO111MODULE=on go build -o gh_dev ./gh

install:
	GO111MODULE=on go install -i ./gh

update: update_all tidy

restore: tidy

tidy:
	GO111MODULE=on go mod tidy

update_all:
	GO111MODULE=on go get -v all

outdated:
	GO111MODULE=on go list -m -u all

release:
	GO111MODULE=on GITHUB_TOKEN=$${GITHUB_TOKEN} goreleaser --rm-dist

test:
	GO111MODULE=on go test -v -race ./...