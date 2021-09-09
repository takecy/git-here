.PHONY: build install update restore tidy update_all outdated

build:
	go build -o gih_dev ./gih

install:
	go install ./gih

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