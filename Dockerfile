FROM golang:alpine3.20 AS builder
WORKDIR /app

COPY . .

RUN go build -o goproxy main.go

FROM golang:alpine3.20
WORKDIR /app

COPY --from=builder /app/goproxy .

ENV PORT 8000

CMD ["./goproxy", "${PORT}"]
