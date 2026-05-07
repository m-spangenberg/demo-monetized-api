FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY . .

WORKDIR /src/services/go-billing
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates libc6-compat

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8081
CMD ["./main"]
