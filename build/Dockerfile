FROM golang:1.23 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix 'static' -o go-template ./cmd/main.go

FROM alpine:latest AS runtime

COPY --from=builder /app/go-template .

ENV CONFIG_PATH="/config/secret.yaml"
VOLUME [ "/config" ]

CMD ["./go-template"]
