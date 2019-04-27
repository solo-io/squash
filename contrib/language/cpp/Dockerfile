FROM gcc:8.3 as build
COPY . /demo
WORKDIR /demo
RUN g++ -std=c++11 -g -o mainDemo main.cpp

# use same alpine as envoy
FROM frolvlad/alpine-glibc as runtime
RUN apk update && apk add --no-cache \ 
    libstdc++
RUN mkdir /usr/local/demo
COPY --from=build /demo/mainDemo /usr/local/demo/mainDemo
WORKDIR /usr/local/demo
CMD ["./mainDemo"]