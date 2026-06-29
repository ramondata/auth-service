# Stage 1
FROM golang:1.26.4-alpine3.24 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service .

# Stage 2
FROM alpine:3.24
WORKDIR /app
COPY --from=builder /app/auth-service .
RUN ls -l /app
EXPOSE 8001
CMD ["./auth-service"]