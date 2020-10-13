FROM ubuntu:18.04 as base

RUN apt-get update
RUN apt-get install --yes gdb build-essential
RUN apt-get install --yes git
RUN apt-get install --yes curl

RUN curl https://storage.googleapis.com/golang/go1.14.9.linux-amd64.tar.gz | tar -C /usr/lib -xz

ENV GOROOT /usr/lib/go
ENV GOPATH /gopath
ENV GOBIN /gopath/bin
ENV PATH $PATH:$GOROOT/bin:$GOPATH/bin
ENV GO111MODULE on

RUN go get github.com/go-delve/delve/cmd/dlv@v1.5.0

FROM base

COPY squash /
ENTRYPOINT ["/squash"]
