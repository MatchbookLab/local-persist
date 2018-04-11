# builder image
FROM golang:1.5 as builder

RUN curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/local-persist
ENV GO15VENDOREXPERIMENT 1

COPY glide.yaml glide.lock ./
RUN glide install -v

COPY ./ ./

RUN go build -o /go/bin/docker-volume-local-persist

# production image
FROM alpine

RUN mkdir -p /var/lib/docker/plugin-data/
COPY --from=builder /go/bin/docker-volume-local-persist .

CMD ["/docker-volume-local-persist"]
#
#
#docker run -d \
#    -v /run/docker/plugins/:/run/docker/plugins/ \
#    -v `pwd`/plugin-data/:/var/lib/docker/plugin-data/ \
#    -v `pwd`:/data/ \
#        cwspear/docker-volume-local-persist-plugin
