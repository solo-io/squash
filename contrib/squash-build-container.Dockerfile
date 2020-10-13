FROM golang:1.14.9-stretch

ENV GO111MODULE on

# Generate dep's cache. this is mainly to save time. so doesn't have to use
# any specific version..
RUN mkdir -p $GOPATH/src/github.com/solo-io && \
    cd $GOPATH/src/github.com/solo-io && \
    git clone https://github.com/solo-io/squash && \
    cd $GOPATH/src/github.com/solo-io/squash && \
    go mod download && \
    rm -rf $GOPATH/src/github.com/solo-io