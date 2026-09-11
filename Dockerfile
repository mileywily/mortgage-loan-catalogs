# Build stage
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the source code
COPY . .

# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mortgage-loan-catalogs ./cmd/server/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/mortgage-loan-catalogs .

# Copy swagger documentation (if available)
COPY --from=builder /app/docs ./docs

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./mortgage-loan-catalogs"]

