FROM golang:1.22-alpine

RUN apk update && \
    apk add --no-cache git curl nano wget htop make golangci-lint && \
    go install github.com/melkeydev/go-blueprint@latest && \
    wget -O - -q https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s v2.20.0