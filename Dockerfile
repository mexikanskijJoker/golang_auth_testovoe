FROM golang:1.23.1-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main cmd/server/main.go

EXPOSE 8000

CMD ["./main"]
