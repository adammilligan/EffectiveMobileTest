FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./

COPY . .

RUN go build -o /bin/subscriptions-api ./cmd/subscriptions-api

FROM alpine:3.21

RUN adduser -D -H -u 10001 app
USER app

COPY --from=builder /bin/subscriptions-api /subscriptions-api
WORKDIR /app

EXPOSE 8080
ENTRYPOINT ["/subscriptions-api"]

