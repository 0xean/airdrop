FROM golang:1.19-buster as builder
ARG GITREF=main

WORKDIR /root
RUN git clone --depth 2 --branch $GITREF https://github.com/arkeonetwork/airdrop.git

WORKDIR /root/airdrop
RUN make install

# final image
FROM ubuntu:noble

WORKDIR /usr/local/bin
COPY --from=builder /go/bin/ .

WORKDIR /root

RUN apt update && apt upgrade -y
RUN apt install -y redis postgresql-client curl wget jq htop vim ssh

COPY ./infra/scripts ./scripts
