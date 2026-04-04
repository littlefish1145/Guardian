FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=latest" -o guardian ./cmd/guardian/

FROM alpine:3.19

RUN apk --no-cache add ca-certificates && \
    mkdir -p /var/log/guardian /app

COPY --from=builder /build/guardian /usr/local/bin/guardian
COPY guardian.minimal.yaml /etc/guardian/guardian.yaml

EXPOSE 8080 9090

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/guardian"]
CMD ["-config", "/etc/guardian/guardian.yaml"]
