# builder image
FROM golang:1.10-alpine as builder

RUN curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/local-persist

RUN set -ex \
    && apk add --no-cache \
    gcc libc-dev curl git \
    && curl https://glide.sh/get | sh

COPY glide.yaml glide.lock ./
RUN glide install -v

COPY ./ ./

RUN go build -o /go/bin/docker-volume-local-persist


# production image
FROM alpine

RUN mkdir -p /var/lib/docker/plugin-data/
COPY --from=builder /go/bin/docker-volume-local-persist /

CMD ["/docker-volume-local-persist"]
