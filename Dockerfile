# Build stage
FROM golang:1.20-alpine AS builder

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    CGO_ENABLED=0 \
    GOOS=linux

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o main .

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
