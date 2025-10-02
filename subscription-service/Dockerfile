# 1. Базовый образ с Go
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . ./

RUN go build -o server ./cmd/server

# -----------------------
# 2. Финальный образ
FROM alpine:latest

RUN apk add --no-cache ca-certificates bash postgresql-client

WORKDIR /app

COPY --from=builder /app/server .
COPY wait-for.sh /wait-for.sh

RUN chmod +x /wait-for.sh

EXPOSE 8080

CMD ["/wait-for.sh", "db:5432", "--", "./server"]
