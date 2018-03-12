#!/bin/sh
docker rm -f server-1
docker build --no-cache --rm=true -t server .
docker run --rm=true -it -p 8888:8888 \
	--name server-1 \
	server
