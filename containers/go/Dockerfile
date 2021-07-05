FROM golang:latest

RUN mkdir -p /bot

WORKDIR /bot

RUN apt update
RUN apt-get -y --no-install-recommends install curl \
    ca-certificates \
    git

ENTRYPOINT "/bot/init.sh"
