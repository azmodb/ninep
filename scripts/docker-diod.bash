#!/bin/bash

# Copyright (c) 2019 The ninep Authors
#
# Permission to use, copy, modify, and distribute this software for any
# purpose with or without fee is hereby granted, provided that the above
# copyright notice and this permission notice appear in all copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
# WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
# MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
# ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
# WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
# ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
# OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

CONTAINER=ninep-client-test
IMAGE=marraison/diod:latest
ADDR=127.0.0.1:5640
EXPORTDIR=/tmp

function usage {
	echo "Usage: docker-diod [OPTIONS]" >&2
	echo  >&2
	echo "The options are:" >&2
	echo >&2
	echo "    -a IP:PORT  set service interface to listen on" >&2
	echo "    -e PATH     export PATH (multiple -e allowed)" >&2
	echo "    -n NAME     container name" >&2
	echo >&2
	exit 2
}

while getopts ":a:e:n:" opt; do
 	case $opt in
 	a) ADDR=${OPTARG};;
	e) exportdir+=(${OPTARG});;
	n) CONTAINER=${OPTARG};;
 	\?) usage;;
	:) usage;;
	esac
done

if [ ${#exportdir[@]} != 0 ]; then
	for dir in ${exportdir[@]}; do
		exportargs+=" --export $dir"
	done
else
	exportargs="--export $EXPORTDIR"
fi

if [ ! $(docker ps --quiet --filter name=${CONTAINER}) ]; then
	docker run --detach --rm --publish ${ADDR}:5640 \
		--name ${CONTAINER} ${IMAGE} \
			${exportargs} \
			--no-auth \
			--debug 3 --logdest stderr


	docker exec -it ${CONTAINER} /bin/sh -c '
		rm -rf /tmp/dir-test && mkdir -p /tmp/dir-test

		mkdir -p /tmp/dir-test/dir1
		mkdir -p /tmp/dir-test/dir1/sub1
		mkdir -p /tmp/dir-test/dir1/sub1/subsub1
		mkdir -p /tmp/dir-test/dir1/sub2
		mkdir -p /tmp/dir-test/dir1/sub3
		mkdir -p /tmp/dir-test/dir2
		mkdir -p /tmp/dir-test/dir3

		touch /tmp/dir-test/dir1/sub1/subsub1/file1
		touch /tmp/dir-test/dir1/file1
		touch /tmp/dir-test/dir1/file2
		touch /tmp/dir-test/file1
		touch /tmp/dir-test/file2
		touch /tmp/dir-test/file3
		touch /tmp/dir-test/file4
		touch /tmp/dir-test/file5
	'
fi
