FROM golang:1.17 as builder

RUN mkdir -p /build

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=linux go build ./cmd/api-gateway/api-gateway.go

FROM golang:1.17

COPY --from=builder /build/api-gateway /app/

WORKDIR /app

CMD ["./api-gateway"]
