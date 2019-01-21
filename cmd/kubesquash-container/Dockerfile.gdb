FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install --yes gdb

ENV DEBUGGER=gdb
COPY kubesquash-container /
ENTRYPOINT ["/kubesquash-container"]
