FROM ubuntu

RUN mkdir -p /bot

WORKDIR /bot

RUN apt update
RUN apt-get -y --no-install-recommends install curl \
    ca-certificates \
    golang \
    git

ENTRYPOINT "/bot/init.sh"
