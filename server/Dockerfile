FROM golang:1.23.5-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/black-lotus

EXPOSE 8080

CMD ["./main"]