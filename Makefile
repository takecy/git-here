.PHONY: build install update restore tidy update_all outdated

build:
	GO111MODULE=on go build -o gih_dev ./gih

install:
	GO111MODULE=on go install -i ./gih

update: update_all tidy

restore: tidy

tidy:
	GO111MODULE=on go mod tidy

update_all:
	GO111MODULE=on go get -v all

outdated:
	GO111MODULE=on go list -m -u all

.PHONY: test
test:
	GO111MODULE=on go test -v -race ./...