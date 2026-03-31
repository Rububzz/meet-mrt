# Stage 1 — build the Go binary
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server *.go

# Stage 2 — run the binary
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY stations.json .
EXPOSE 8080
CMD ["./server"]