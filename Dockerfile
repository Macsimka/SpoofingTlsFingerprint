FROM golang:alpine3.20 AS builder
WORKDIR /app

COPY . .

RUN go build -o myapp main.go

FROM golang:alpine3.20
WORKDIR /app

COPY --from=builder /app/myapp .

CMD ["./myapp"]
