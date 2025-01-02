FROM golang:1.22-alpine

WORKDIR /app

# Install netcat for wait-for-it script
RUN apk add --no-cache netcat-openbsd

# Copy wait-for-it script
COPY scripts/wait-for-it.sh /usr/local/bin/wait-for-it.sh
RUN chmod +x /usr/local/bin/wait-for-it.sh

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/main.go

# Expose port 8080 to the outside
EXPOSE 8080

# Use wait-for-it script to wait for postgres
CMD ["sh", "-c", "wait-for-it.sh postgres 5432 && ./main"]
