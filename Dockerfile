FROM golang

# Work directory inside the container
WORKDIR /go/src/github.com/ailinykh/pullanusbot

# Copy the local package files to the container's workspace.
ADD . .

# Install dependencies
RUN curl -L -s https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o /go/bin/dep \
    chmod +x /go/bin/dep
    
RUN /go/bin/dep ensure

# Build the pullanusbot inside the container.
RUN go build && go install

# Mount data directory
VOLUME [ "data" ]

# Run the pullanusbot by default when the container starts.
ENTRYPOINT /go/bin/pullanusbot
