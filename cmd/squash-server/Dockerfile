FROM alpine:3.5

COPY squash-server /squash-server

ENTRYPOINT ["/squash-server", "--cluster=kube", "--host=0.0.0.0", "--port=8080"]

EXPOSE 8080
