FROM golang:1.12-alpine as builder

WORKDIR /go/src/ninep
COPY . .

ENV GO111MODULE=on
RUN set -eux; apk add --no-cache --virtual .build-deps \
        git; \
    CGO_ENABLED=0 GOOS=linux go install -a -v \
        -installsuffix cgo \
        -ldflags '-extldflags "-static"' ./...; \
    apk del .build-deps

FROM alpine:3.10

COPY --from=builder /go/bin/ninepd /bin/ninepd

EXPOSE 5640
VOLUME /export

ENTRYPOINT ["/bin/ninepd"]
CMD ["-addr", ":5640", "-export", "/export"]
