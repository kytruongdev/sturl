# =========================
# Stage 1: builder
# =========================
FROM golang:1.25 AS builder

WORKDIR /app

# Cache go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/server ./cmd/server && \
    go build -o /bin/producer ./cmd/producer && \
    go build -o /bin/consumer ./cmd/consumer

# =========================
# Stage 2: runtime
# =========================
FROM gcr.io/distroless/base-debian12 AS runtime

WORKDIR /app

# Copy binaries only
COPY --from=builder /bin/server /app/server
COPY --from=builder /bin/producer /app/producer
COPY --from=builder /bin/consumer /app/consumer

# Non-root user
USER nonroot:nonroot

# Default command (override in compose)
CMD ["/app/server"]
