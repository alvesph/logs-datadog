FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /go-http-service

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /go-http-service /app/go-http-service

EXPOSE 8080

CMD ["/app/go-http-service"]
