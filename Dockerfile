# Just for building
FROM golang:1.12-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/elastifeed/es-scraper

# Enable go Modules
ENV GO111MODULE=on

# Copy source files
COPY . .

# Fetch deps dependencies
RUN go get -d -v ./...

# Build and Install executables
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go && mkdir -p /go/bin/ && mv main /go/bin/es-scraper

# Use the chromedp headless image as described under https://github.com/chromedp/chromedp#frequently-asked-questions
FROM ubuntu:18.04

LABEL maintainer="Matthias Riegler <me@xvzf.tech>"

RUN apt-get update && apt-get upgrade -y && apt-get install libxss1 libappindicator1 libindicator7
 && rm -rf /var/lib/apt/lists/*

# Try installing the actual chrome client
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && apt install ./google-chrome*.deb


COPY --from=builder /go/bin/es-scraper /go/bin/es-scraper

# Fixed port
ENV API_BIND=":9090"

# Set path
# ENV PATH $PATH:/headless-shell

ENTRYPOINT ["/go/bin/es-scraper"]

EXPOSE 8080
