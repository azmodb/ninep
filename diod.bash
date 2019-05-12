#!/bin/bash
CONTAINER=ninep-client-test
IMAGE=marraison/diod:latest

if [ ! $(docker ps --quiet --filter name=${CONTAINER}) ]; then
	docker run --detach --rm --publish 127.0.0.1:5640:5640 \
		--name ${CONTAINER} ${IMAGE} \
			--export /tmp --no-auth \
			--debug 3 --logdest stderr
fi
