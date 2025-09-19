FROM docker.io/golang:alpine as builder
WORKDIR /app

# copy go.mod and go.sum separately to avoid unnecessarily redownloading dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" .

FROM docker.io/alpine:latest
WORKDIR /app

COPY --from=builder /app/lancache-dns-generator /app

EXPOSE 3333

CMD ["/app/lancache-dns-generator"]
