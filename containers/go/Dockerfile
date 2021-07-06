FROM golang:latest

RUN mkdir -p /bot

WORKDIR /bot

RUN apt update
RUN apt-get -y --no-install-recommends install curl \
    ca-certificates \
    ssh \
    build-essential \
    git

RUN chmod 777 /bot/cache -R

RUN mkdir -p /root/.ssh
ADD ./id_rsa.pub /root/.ssh/id_rsa.pub
ADD ./knows_hosts /root/.ssh/knows_hosts
RUN chmod 764 -R /root/.ssh
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

ENTRYPOINT "/bot/init.sh"
