export GO111MODULE=on

all:build

build:
	go mod tidy
	go build
