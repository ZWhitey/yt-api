FROM golang:1.20.2-alpine3.17 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./api.go ./

RUN go mod download

RUN go build -o main .

FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/main ./

CMD ["./main"]
