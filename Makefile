-include .env

.PHONY: run build clean

run: build
	./pullanusbot

build: clean *.go
	go build .

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic \
		pullanusbot/config \
		pullanusbot/converter \
		pullanusbot/faggot \
		pullanusbot/info \
		pullanusbot/interfaces \
		pullanusbot/link \
		pullanusbot/twitter \
		pullanusbot/utils \
		pullanusbot/youtube \

clean:
	rm -f pullanusbot