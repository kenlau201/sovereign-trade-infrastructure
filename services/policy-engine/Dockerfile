FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY services/policy-engine/go.mod .
COPY services/policy-engine/go.sum . 2>/dev/null || true

# Download dependencies
RUN go mod download

# Copy source code
COPY services/policy-engine/ .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o policy-engine main.go

# Runtime image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/policy-engine .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./policy-engine"]
