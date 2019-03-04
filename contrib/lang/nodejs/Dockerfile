# From: https://kubernetes.io/docs/tutorials/hello-minikube/
# minikube/Dockerfile 

FROM node:11-alpine
EXPOSE 8080 9229
COPY server.js .
CMD node --inspect=127.0.0.1:9229 server.js