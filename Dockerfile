FROM golang:stretch as builder
WORKDIR /go/src/github.com/ailinykh/pullanusbot
# cache dependencies first
COPY go.mod go.sum ./
RUN go mod download
# now build
ADD . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"'

FROM jrottenberg/ffmpeg:4.4-alpine313
RUN apk update && apk add tzdata python3 --no-cache && \
    wget https://yt-dl.org/downloads/latest/youtube-dl -O /usr/local/bin/youtube-dl && \
    chmod a+rx /usr/local/bin/youtube-dl && \
    ln -s /usr/bin/python3 /usr/bin/python
COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/pullanusbot /usr/local/bin/pullanusbot
WORKDIR /usr/local/share
VOLUME [ "/usr/local/share/pullanusbot-data" ]
ENTRYPOINT pullanusbot