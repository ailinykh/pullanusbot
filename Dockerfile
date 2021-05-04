FROM golang:stretch as builder
WORKDIR /go/src/github.com/ailinykh/pullanusbot
# cache dependencies first
COPY go.mod go.sum ./
RUN go mod download
# now build
ADD . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"'

FROM jrottenberg/ffmpeg:4.1-alpine
RUN apk update && apk add tzdata python3 supervisor openssh --no-cache && \
    ssh-keygen -f /etc/ssh/ssh_host_rsa_key -N '' -t rsa && \
    ssh-keygen -f /etc/ssh/ssh_host_dsa_key -N '' -t dsa && \
    wget https://yt-dl.org/downloads/latest/youtube-dl -O /usr/local/bin/youtube-dl && chmod a+rx /usr/local/bin/youtube-dl

COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/pullanusbot /usr/local/bin/pullanusbot
COPY bin/telegram-bot-api /usr/local/bin/telegram-bot-api
COPY supervisord.conf /etc/supervisord.conf

WORKDIR /usr/local/share
VOLUME [ "pullanusbot-data" ]
ENTRYPOINT supervisord -c /etc/supervisord.conf