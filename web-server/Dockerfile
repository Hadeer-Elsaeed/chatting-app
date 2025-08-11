FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod file and source code
COPY go.mod ./
COPY *.go ./

# Download dependencies and tidy (this will create go.sum and update go.mod)
RUN go mod tidy
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o web-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/web-server .

# Create uploads directory
RUN mkdir -p uploads

# Expose port
EXPOSE 8080

# Command to run
CMD ["./web-server"]
