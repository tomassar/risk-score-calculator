FROM golang:1.23 AS builder

RUN apt-get update && apt-get install -y \
    gcc \
    sqlite3 \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest

RUN templ generate

RUN CGO_ENABLED=1 go build -o main cmd/server/main.go

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/web/static ./web/static

RUN mkdir -p data

EXPOSE 8080

CMD ["./main"] 