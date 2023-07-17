FROM golang:stretch as builder
WORKDIR /go/src/github.com/ailinykh/pullanusbot
# cache dependencies first
COPY go.mod go.sum ./
RUN go mod download
# now build
ADD . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"'

FROM jrottenberg/ffmpeg:5.1-alpine313
RUN apk update && apk add tzdata python3 --no-cache && \
    wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp && \
    ln -s /usr/bin/python3 /usr/bin/python
COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/pullanusbot /usr/local/bin/pullanusbot
WORKDIR /usr/local/share
VOLUME [ "/usr/local/share/pullanusbot-data" ]
ENTRYPOINT pullanusbot