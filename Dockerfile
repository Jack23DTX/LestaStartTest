FROM golang:1.23-alpine AS builder

WORKDIR /web-app

COPY .env ./
COPY go.mod go.sum ./

RUN go mod download

COPY internal ./internal
COPY cmd ./cmd

RUN go build -o /web-app/LestaStartTest ./cmd/

COPY internal/templates ./internal/templates

EXPOSE 8080

CMD ["/web-app/LestaStartTest"]