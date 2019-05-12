#!/bin/bash

CONTAINER=ninep-client-test
IMAGE=marraison/diod:latest
ADDR=127.0.0.1:5640
EXPORTDIR=/tmp

if [ ! $(docker ps --quiet --filter name=${CONTAINER}) ]; then
	docker run --detach --rm --publish ${ADDR}:5640 \
		--name ${CONTAINER} ${IMAGE} \
			--export ${EXPORTDIR} \
			--no-auth \
			--debug 3 --logdest stderr
fi
