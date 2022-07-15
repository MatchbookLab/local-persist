FROM golang:1.18 as builder

WORKDIR /build
COPY . .
RUN go build -o local-persist

# generate clean, final image for end users
FROM alpine:3.11.3
COPY --from=builder /build/local-persist .
RUN mkdir -p /run/docker/plugins
# executable
ENTRYPOINT [ "./local-persist" ]