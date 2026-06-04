# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o snip ./cmd/server

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
COPY --from=builder /app/snip .
RUN mkdir -p /app/data
EXPOSE 53524
ENV SNIP_HOST=0.0.0.0
ENV SNIP_PORT=53524
ENV SNIP_DB_PATH=/app/data/snip.db
VOLUME ["/app/data"]
CMD ["./snip"]
