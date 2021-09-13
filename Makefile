.PHONY: build install update restore tidy update_all outdated

build:
	go build -o gih_dev -ldflags "-X main.version='0.0.1-test'" ./cmd

install:
	go install ./cmd

update: update_all tidy

restore: tidy

tidy:
	go mod tidy

update_all:
	go get -u -v ./...

outdated:
	go list -m -u all

.PHONY: test
test:
	go test -v -race ./...