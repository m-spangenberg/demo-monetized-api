FROM golang:1.26.2-alpine AS builder

WORKDIR /src

COPY . .

WORKDIR /src/services/go-validator
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
