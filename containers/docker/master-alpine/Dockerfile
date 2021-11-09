FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://git.taiyue.io/pist/go-pist && \
  (cd pist && make gpist) && \
  cp pist/build/bin/gpist /gpist && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /pist && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gpist"]
