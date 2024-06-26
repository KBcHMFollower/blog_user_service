FROM golang:1.22.4-alpine

WORKDIR /app

COPY . .

RUN go build -o auth cmd/auth/main.go

CMD ["./auth --config=config/prod.yaml"]