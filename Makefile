include .env

.PHONY: run

run: build
	./pullanusbot

build: *.go
	go build .