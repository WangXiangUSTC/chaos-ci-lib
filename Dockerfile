FROM golang:latest

LABEL maintainer="LitmusChaos"
RUN apt-get update && apt-get install -y git && \
    apt-get install -y ssh && \
    apt install ssh rsync

ENV CGO_ENABLED 0
ENV GO111MODULE=on
RUN export GOPATH=$HOME/go
RUN export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
ARG KUBECTL_VERSION=1.17.0

RUN apt-get update && apt-get install -y python-pip && \
    pip install --upgrade pip && \
    pip install --upgrade setuptools

RUN git clone https://github.com/helm/helm.git &&\
    cd helm && make
    
RUN apt-get update -y \
    && apt-get install -y \
    python3 \
    python3-pip
RUN pip3 install --upgrade pip
RUN python3 -m pip install pygithub

ADD https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

COPY pkg ./pkg
COPY tests ./tests  
COPY types ./types
COPY go.mod ./go.mod
COPY go.sum ./go.sum
COPY vendor/ ./vendor
COPY app ./app
