FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/LestaStartTest ./cmd/

# Финальный образ
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/LestaStartTest .

COPY --from=builder /app/static ./static

COPY .env .

RUN mkdir uploads

RUN apk add --no-cache libc6-compat

EXPOSE 8080

CMD ["./LestaStartTest"]