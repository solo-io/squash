FROM golang:1.9.2-stretch

# get dep vendoring tool
RUN go get -u github.com/golang/dep/cmd/dep

# Generate dep's cache. this is mainly to save time. so doesn't have to use
# any specific version..
RUN mkdir -p $GOPATH/src/github.com/solo-io && \
    cd $GOPATH/src/github.com/solo-io && \
    git clone https://github.com/solo-io/squash && \
    cd $GOPATH/src/github.com/solo-io/squash && \
    dep ensure && \
    rm -rf $GOPATH/src/github.com/solo-io