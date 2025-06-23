FROM golang:1.21-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest

RUN templ generate

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/web/static ./web/static

RUN mkdir -p data

EXPOSE 8080

CMD ["./main"] 