FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /monitoring .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /monitoring /monitoring
EXPOSE 9091
ENTRYPOINT ["/monitoring"]
