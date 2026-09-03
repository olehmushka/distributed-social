### Compile stage
FROM golang:1.27.1-alpine AS build-env
RUN apk add --no-cache build-base make git

ADD . /dockerbuild
WORKDIR /dockerbuild

ENV GOEXPERIMENT=loopvar
RUN go build -tags timetzdata -o /search ./cmd/search

### Run stage
FROM alpine:3.24

RUN apk add --no-cache --update dumb-init ca-certificates wget && \
    addgroup -S app && adduser -S app -G app

ENTRYPOINT ["dumb-init", "--"]

WORKDIR /
COPY --from=build-env /search /

ENV GODEBUG=netdns=go
ENV TZ=Etc/UTC

USER app

EXPOSE 9012

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O- http://localhost:9012/ping || exit 1

CMD ["/search", "run"]
