build:
	go build -o gs_dev ./gs

build_c:
	go get -u github.com/mitchellh/gox
	gox ./gs
