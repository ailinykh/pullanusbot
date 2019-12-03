FROM golang:stretch as builder
WORKDIR /go/src/github.com/ailinykh/pullanusbot
# cache dependencies first
COPY go.mod go.sum ./
RUN go mod download
# now build
ADD . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"'

FROM jrottenberg/ffmpeg:4.1-alpine
WORKDIR /go/bin
COPY --from=builder /go/src/github.com/ailinykh/pullanusbot/pullanusbot .
VOLUME [ "data" ]
ENTRYPOINT /go/bin/pullanusbot