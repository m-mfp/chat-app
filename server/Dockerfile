# Stage 1: Build
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o server

# Stage 2: Run
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8000

CMD ["./server"]
