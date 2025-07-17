FROM golang:1.24.4-alpine

WORKDIR /app

# Copy and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN go build -o server

# Expose port
EXPOSE 8000

# Run the app
CMD ["./server"]