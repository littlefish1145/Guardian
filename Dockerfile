FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o guardian ./cmd/guardian/

FROM nginx:alpine

COPY --from=builder /build/guardian /usr/local/bin/
COPY --from=builder /build/guardian.yaml /etc/guardian/guardian.yaml

# RUN apk add --no-cache openrc docker-cli python3
# RUN apk add --no-cache python3
EXPOSE 9090

ENTRYPOINT ["/usr/local/bin/guardian", "-config", "/etc/guardian/guardian.yaml"]
