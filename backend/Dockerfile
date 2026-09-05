FROM golang:1.23-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -H -u 10001 appuser

COPY --from=builder --chown=appuser:appuser /out/api /app/api

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/api"]
