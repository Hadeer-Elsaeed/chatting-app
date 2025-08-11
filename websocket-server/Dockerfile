FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files and source code
COPY go.mod ./
COPY go.sum* ./
COPY . .

# Download dependencies and tidy up (this will analyze the source code and update dependencies)
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o websocket-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/websocket-server .

# Expose port
EXPOSE 8081

# Command to run
CMD ["./websocket-server"]
