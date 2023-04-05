FROM golang:1.20 as builder

WORKDIR /build
COPY . .
RUN env CGO_ENABLED=0 go build -o local-persist

# generate clean, final image for end users
FROM alpine

COPY --from=builder /build/local-persist /usr/bin/local-persist

RUN mkdir -p /run/docker/plugins

# executable
ENTRYPOINT ["/usr/bin/local-persist"]
