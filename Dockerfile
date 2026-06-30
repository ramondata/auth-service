# Etapa 1: build
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auth-service .

# Etapa 2: imagem final apenas com o binário
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/auth-service /app/auth-service

USER nonroot:nonroot

EXPOSE 8001

ENTRYPOINT ["/app/auth-service"]