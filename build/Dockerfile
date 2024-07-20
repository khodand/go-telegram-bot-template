FROM golang:1.22 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix 'static' -o telegram-bot ./cmd/main.go

FROM alpine:latest AS runtime

COPY --from=builder /app/telegram-bot .

ENV CONFIG_PATH="/config/secret.yaml"
VOLUME [ "/config" ]

CMD ["./telegram-bot"]
