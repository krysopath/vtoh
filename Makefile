GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_TAG := $(shell git describe --abbrev=0 HEAD 2>/dev/null)
LD_FLAGS := '-s -w -X main.gitTag=$(GIT_TAG) -X main.gitRef=$(GIT_SHA)'

/usr/bin/token-helper: build/bin/token-helper
	sudo cp build/bin/token-helper /usr/bin/token-helper

build/bin/token-helper:  *.go go.* deps.txt
	go build -ldflags=$(LD_FLAGS) -o build/bin/token-helper

deps.txt: go.mod go.sum
	go mod tidy
	go get
	go mod graph > deps.txt
