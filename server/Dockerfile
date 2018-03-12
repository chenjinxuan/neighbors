FROM golang:1.7.6
MAINTAINER chenjinxuan <jinxuanchen666@163.com>
COPY . /go/src/geohash/server
RUN go install geohash/server
ENTRYPOINT ["/go/bin/server"]
