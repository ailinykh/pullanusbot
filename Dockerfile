FROM golang

# Work directory inside the container
WORKDIR /go/src/github.com/ailinykh/pullanusbot

# Copy the local package files to the container's workspace.
ADD . .

# Install dependencies
RUN go get gopkg.in/tucnak/telebot.v2

# Build the pullanusbot inside the container.
RUN go build && go install

# Mount data directory
VOLUME [ "data" ]

# Run the pullanusbot by default when the container starts.
ENTRYPOINT /go/bin/pullanusbot
