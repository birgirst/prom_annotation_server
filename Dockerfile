FROM       ubuntu:14.04
MAINTAINER Birgir Stefansson <birgirst@cxense.com>
EXPOSE     9119
ENTRYPOINT [ "./prom_annotation_server", "--storage=local:/var/annotations/annotations.db" ]
WORKDIR    /usr/src/prom_annotation_server
COPY       . /usr/src/prom_annotation_server

RUN apt-get -y install wget \
    && apt-get -y install git \
    && wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz \
    && sudo tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz \
    && export PATH=$PATH:/usr/local/go/bin \
    && export GOPATH=/usr/share/go \
    && go get -u github.com/boltdb/bolt/... \
    && go get -u github.com/dancannon/gorethink/... \
    && go get -u github.com/prometheus/client_golang/prometheus/... \
    && go build
