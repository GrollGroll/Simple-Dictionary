FROM golang:1.22

WORKDIR /app

COPY . .

CMD ["go", "run", "cmd/main.go"]
