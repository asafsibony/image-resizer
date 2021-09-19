FROM golang:1.17 as builder

RUN mkdir -p /build

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=linux go build ./cmd/image-resizer-ms/image-resizer-ms.go

FROM golang:1.17

COPY --from=builder /build/image-resizer-ms /app/

WORKDIR /app

CMD ["./image-resizer-ms"]
