FROM golang:1.25.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o notification-service ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/notification-service .
EXPOSE 8000
CMD ["./notification-service"]