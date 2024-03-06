FROM golang:1.22-bookworm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["go", "run", "main.go"]
