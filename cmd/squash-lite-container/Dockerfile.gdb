FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install --yes gdb

ENV DEBUGGER=gdb
COPY squash-lite-container /
ENTRYPOINT ["/squash-lite-container"]
