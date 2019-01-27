FROM golang:1.5 as builder

RUN curl -sSL https://glide.sh/get | sed 's+get TAG https://glide.sh/version+TAG=v0.12.3+' | sh

WORKDIR /go/src/local-persist
ENV GO15VENDOREXPERIMENT 1

COPY glide.yaml glide.lock ./
RUN glide install -v

COPY ./ ./

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM
ENV CGO_ENABLED=0
RUN go build -o bin/local-persist -v

RUN mkdir -p /var/lib/docker/plugin-data/


FROM scratch

COPY --from=builder /go/src/local-persist/bin/local-persist /bin/local-persist
COPY --from=builder /var/lib/docker/plugin-data/ /var/lib/docker/plugin-data/

# TODO check and optimze
# WIthout this line starting container fails with
# > docker run --rm -it lp
# Starting...       Found 0 volumes on startup
# open /etc/group: no such file or directory
COPY --from=builder /etc/group /etc/group

ENTRYPOINT ["/bin/local-persist"]
