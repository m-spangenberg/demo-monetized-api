# Use a multi-stage build to include the common package
FROM golang:1.26.2-alpine AS builder

# Install build dependencies for cgo (required for sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

# Copy the entire workspace root to access 'pkg/common'
COPY . .

# Build the specific service
WORKDIR /src/services/go-api
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main main.go

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates libc6-compat

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8082
CMD ["./main"]
