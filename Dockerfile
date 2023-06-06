FROM golang:bullseye

RUN apt update \
 && apt install -yq \
    git \
 && git config --global user.email foo@bar.com

WORKDIR /project
