FROM golang:1.23-bookworm as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /main cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /main .

EXPOSE 8080

CMD ["./main"]
