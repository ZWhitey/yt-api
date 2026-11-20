FROM golang:1.27-alpine3.23 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY ./cmd ./cmd

COPY ./internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o main ./cmd

FROM scratch

WORKDIR /app

ARG VERSION=unknown
ENV VERSION=${VERSION}

COPY --from=builder /app/main ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["./main"]
