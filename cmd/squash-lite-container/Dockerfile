FROM golang:1.10.1-stretch as builder

RUN mkdir -p $GOPATH/src/github.com/derekparker/ && cd  $GOPATH/src/github.com/derekparker/ && git clone https://github.com/derekparker/delve/
RUN cd $GOPATH/src/github.com/derekparker/delve/ && git checkout v1.0.0
RUN cd $GOPATH/src/github.com/derekparker/delve/cmd/dlv && go install


FROM golang:1.10.1-stretch

COPY --from=builder $GOPATH/bin/dlv $GOPATH/bin/

COPY squash-lite-container /
ENTRYPOINT ["/squash-lite-container"]
