FROM alpine:3.5

RUN apk update && \
    apk add vim && \
    rm -rf /var/cache/apk/*

ENV VIM_VERSION 2.5.0

EXPOSE 53/udp

ENTRYPOINT ["/usr/vim"]