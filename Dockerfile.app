FROM golang:1.24-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./                    
COPY internal/ ./internal/
COPY cmd/ ./cmd/
COPY config/ ./config/

COPY app/migrations ./migrations

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/main /main
EXPOSE 7490
CMD ["/main", "http"]