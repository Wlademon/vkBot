FROM ubuntu

run mkdir -p /bot

WORKDIR /bot

run apt update
run apt-get -y --no-install-recommends install curl \
    ca-certificates


ENTRYPOINT "/bot/init.sh"
