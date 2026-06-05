# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o snip ./cmd/server

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /app/snip .
RUN mkdir -p /app/data && adduser -D -H -h /app snip && chown -R snip:snip /app
USER snip
EXPOSE 53524
ENV SNIP_HOST=0.0.0.0
ENV SNIP_PORT=53524
ENV SNIP_DB_PATH=/app/data/snip.db
VOLUME ["/app/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:53524/healthz || exit 1
CMD ["./snip"]
