# Build Stage
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/eslutz/forwardarr/pkg/version.Version=${VERSION} -X github.com/eslutz/forwardarr/pkg/version.Commit=${COMMIT} -X github.com/eslutz/forwardarr/pkg/version.Date=${DATE}" \
    -o forwardarr ./cmd/forwardarr

# Runtime Stage
FROM alpine:3.23
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
# Create non-root user
RUN addgroup -g 1000 forwardarr && adduser -u 1000 -G forwardarr -D forwardarr

COPY --from=builder /app/forwardarr /app/forwardarr

USER forwardarr
EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

ENTRYPOINT ["/app/forwardarr"]
