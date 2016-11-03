FROM debian

ENV VERSION v1.3.0
ENV ARCH amd64

ADD https://github.com/CWSpear/local-persist/releases/download/${VERSION}/local-persist-linux-${ARCH} /usr/bin/docker-volume-local-persist
RUN chmod +x /usr/bin/docker-volume-local-persist

ADD https://github.com/Yelp/dumb-init/releases/download/v1.1.1/dumb-init_1.1.1_amd64 /usr/bin/dumb-init
RUN chmod +x /usr/bin/dumb-init

RUN mkdir -p /var/lib/docker/plugin-data/

CMD ["dumb-init", "/usr/bin/docker-volume-local-persist"]
