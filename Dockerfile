# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 go build -o /build/server ./cmd/server

# Runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /build/server .

EXPOSE 8080

CMD ["./server"]
