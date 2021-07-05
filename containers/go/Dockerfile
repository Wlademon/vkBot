FROM golang:latest

RUN mkdir -p /bot

WORKDIR /bot

RUN apt update
RUN apt-get -y --no-install-recommends install curl \
    ca-certificates \
    ssh \
    build-essential \
    git

RUN mkdir -p ~/.ssh
RUN chmod 751 -R ~/.ssh
ADD id_rsa.pub ~/.ssh/id_rsa.pub
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

ENTRYPOINT "/bot/init.sh"
