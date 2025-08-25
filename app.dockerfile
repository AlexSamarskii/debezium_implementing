# Dockerfile.app
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY config.yaml ./config.yaml
COPY main.go ./main.go
COPY internal ./internal
COPY cmd ./cmd
COPY config ./config
COPY migrations ./migrations
COPY .env ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin ./main.go

# Финальный образ
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/migrate .
COPY --from=builder /app/.bin .
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/.env .

RUN apk add --no-cache tzdata libc6-compat
ENV TZ=Europe/Moscow
