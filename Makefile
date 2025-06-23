.PHONY: build run test clean templ dev

build: templ
	go build -o bin/server cmd/server/main.go

run: templ
	go run cmd/server/main.go

dev:
	templ generate --watch --proxy="http://localhost:8080" --cmd="go run cmd/server/main.go"

templ:
	templ generate

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf data/

deps:
	go mod download
	go install github.com/a-h/templ/cmd/templ@latest

fmt:
	go fmt ./...
	templ fmt .

lint:
	golangci-lint run

docker-build:
	docker build -t risk-calculator .

docker-run:
	docker run -p 8080:8080 risk-calculator 