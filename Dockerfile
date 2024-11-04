FROM --platform=linux/amd64 golang:1.22 as builder
WORKDIR /go/src/github.com/ailinykh/pullanusbot
ADD . .
RUN go mod download
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -v -a -installsuffix cgo -ldflags '-extldflags "-static"' ./cmd/bot

FROM jrottenberg/ffmpeg:6.1-alpine
LABEL maintainer="Anton Ilinykh <anthonyilinykh@gmail.com>"
RUN apk update && apk add tzdata python3 --no-cache && \
    wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -O /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp && \
    ln -s /usr/bin/python3 /usr/bin/python
COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/bot /usr/local/bin/bot
WORKDIR /usr/local/share
VOLUME [ "/usr/local/share/pullanusbot-data" ]
ENTRYPOINT bot