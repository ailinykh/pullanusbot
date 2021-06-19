-include .env

.PHONY: test run build clean

run: build
	./pullanusbot

test:
	go test ./... -coverprofile=cover.txt -race

build: clean *.go
	go build .

clean:
	rm -f pullanusbot