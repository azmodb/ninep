FROM golang:1.12.7-alpine

WORKDIR /go/src/ninep
COPY . .

ENV GO111MODULE=on
RUN set -eux; apk add --no-cache \
		git \
		gcc \
		libc-dev \
		; \
	go get -d ./...

ENTRYPOINT ["go"]
CMD ["test", "./..."]
