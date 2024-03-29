.PHONY: build install update restore tidy update_all

build:
	go build -o gih_dev -ldflags "-X main.version='0.0.1-test'" ./gih

install:
	go install ./gih

update: update_all tidy

lint:
	golangci-lint run

restore: tidy

tidy:
	go mod tidy

update_all:
	go get -u -v ./...


.PHONY: test
test:
	go test -v -race ./...