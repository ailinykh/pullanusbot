-include .env

.PHONY: test run build clean

all: build run

run:
	./pullanusbot

test:
	GO_ENV=testing go test ./... -v -coverprofile=coverage.txt -race -covermode=atomic

build: clean *.go
	go build .

clean:
	rm -f pullanusbot