# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/migrate ./cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrate .
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

RUN mkdir -p uploads

EXPOSE 3000

ENTRYPOINT ["./docker-entrypoint.sh"]
