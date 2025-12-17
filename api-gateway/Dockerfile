# =========================
# Builder
# =========================
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/api-gateway ./cmd/server

# =========================
# Runtime
# =========================
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /bin/api-gateway /app/api-gateway

USER nonroot:nonroot

CMD ["/app/api-gateway"]
