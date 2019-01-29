FROM golang:1.5 as builder

WORKDIR /go/src/local-persist
ENV GO15VENDOREXPERIMENT 1

RUN curl -sSL https://glide.sh/get | sed 's+get TAG https://glide.sh/version+TAG=v0.12.3+' | sh
COPY glide.lock glide.yaml ./
RUN glide install -v

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM
ENV CGO_ENABLED=0
COPY driver.go main.go ./
RUN go build -o bin/local-persist -v

RUN mkdir -p /tmp/plugin-data/ && \
    echo "root:x:0:" > /tmp/group


FROM scratch

COPY --from=builder /go/src/local-persist/bin/local-persist /bin/local-persist
COPY --from=builder /tmp/plugin-data/ /var/lib/docker/plugin-data/
COPY --from=builder /tmp/group /etc/group

ENTRYPOINT ["/bin/local-persist"]
