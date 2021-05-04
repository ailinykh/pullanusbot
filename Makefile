.PHONY: test run build clean

run: build
	./pullanusbot

test:
	go test ./... -coverprofile=cover.txt

build: clean *.go
	go build .

clean:
	rm -f pullanusbot