-include .env

.PHONY: all serve kill run test build clean restart

APP ?= bin/pullanusbot
PID = $(APP).pid
GO_FILES = $(wildcard *.go)

all: serve

before:
	@echo "🛠 rebuilding an app..."

serve: run
	@fswatch -x -o --event Created --event Updated --event Renamed -r -e '.*' -i '\.go$$'  . | xargs -n1 -I{}  make restart || make kill

$(APP): $(GO_FILES)
	@go build $? -o $@

kill:
	@kill `cat $(PID)` || true

run: build
	@$(APP) & echo $$! > $(PID)

test:
	GO_ENV=testing go test ./... -v -coverprofile=coverage.txt -race -covermode=atomic

build: $(GO_FILES)
	@go build -o $(APP) .

clean:
	rm -f $(APP)

restart: kill before build run