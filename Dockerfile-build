FROM golang:1.5

RUN curl https://glide.sh/get | sh

WORKDIR $GOPATH/src/local-persist
ENV GO15VENDOREXPERIMENT 1

COPY glide.yaml glide.lock ./
RUN glide install -v

COPY ./ ./

CMD ["make", "binaries"]
