FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o load_balancer cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/load_balancer .
COPY configs/config.json .
ENV CONFIG_PATH=/app/config.json
EXPOSE 8080
CMD ["./load_balancer"]