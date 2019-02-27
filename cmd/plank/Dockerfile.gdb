FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install --yes gdb

ENV DEBUGGER=gdb
COPY plank /
ENTRYPOINT ["/plank"]
