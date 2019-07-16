# Just for building
FROM golang:1.12-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/elastifeed/es-scraper

# Enable go Modules
ENV GO111MODULE=on

# Copy source files
COPY . .

# Fetch deps dependencies
RUN go get -u -v ./...

# Build and Install executables
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go && mkdir -p /go/bin/ && mv main /go/bin/es-scraper

LABEL maintainer="Matthias Riegler <me@xvzf.tech>"

# Fixed port
ENV API_BIND=":9090"

ENTRYPOINT ["/go/bin/es-scraper"]

EXPOSE 8080
