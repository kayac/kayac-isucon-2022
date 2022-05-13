FROM golang:1.18.1-bullseye AS builder

COPY . /tmp/bench/
WORKDIR /tmp/bench/
RUN make bench httpd

FROM debian:bullseye-slim

RUN apt-get update && apt-get -y upgrade && apt-get -y install ca-certificates curl
RUN mkdir -p /home/isucon/bench
COPY --from=builder /tmp/bench/httpd /usr/local/bin/httpd
COPY --from=builder /tmp/bench/bench /home/isucon/bench/bench
COPY --from=builder /tmp/bench/data/ /home/isucon/bench/data/
WORKDIR /home/isucon/bench

ENTRYPOINT ["/usr/local/bin/httpd"]
