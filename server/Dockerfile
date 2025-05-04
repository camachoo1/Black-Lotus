FROM golang:1.23.5-alpine

# Add build tools and clean cache in the same layer to reduce image size
RUN apk add --no-cache git && \
    mkdir -p /tmp/go-cache /tmp/go-mod-cache

# Set Go's cache directories to temp locations
ENV GOCACHE=/tmp/go-cache
ENV GOMODCACHE=/tmp/go-mod-cache

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Use trimpath to reduce size and avoid path issues
RUN go build -trimpath -o main ./cmd/black-lotus

EXPOSE 8080

CMD ["./main"]