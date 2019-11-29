FROM golang:stretch as builder

RUN apt-get install curl -y && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/ailinykh/pullanusbot
ENV PATH="/go/src/github.com/ailinykh/pullanusbot:${PATH}"
ADD . .
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM jrottenberg/ffmpeg:4.1-alpine
WORKDIR /go/bin
COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/pullanusbot .
ENV PATH="/go/bin:${PATH}"
VOLUME [ "data" ]
ENTRYPOINT /go/bin/pullanusbot
