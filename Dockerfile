FROM golang:1.20.2-alpine3.17 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./cmd ./cmd

COPY ./internal ./internal

RUN go build -o main ./cmd

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/main ./

CMD ["./main"]
