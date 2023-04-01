FROM golang:1.19 as builder

WORKDIR /build
COPY . .
RUN env CGO_ENABLED=0 go build -o local-persist

# generate clean, final image for end users
FROM alpine

COPY --from=builder /build/local-persist local-persist

RUN mkdir -p /run/docker/plugins /state /docker-data

# executable
ENTRYPOINT [ "/local-persist" ]
