# Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o crm-lite main.go

EXPOSE 8080

CMD ["./crm-lite"]