#!/bin/bash -x -e

# minikuberestart

 # kubectl create ns demo

 sleep 2

squashctl -h

squashctl deploy demo \
--demo-id go-java \
--demo-namespace1 demo \
--demo-namespace2 demo

squashctl deploy agent
squashctl # (interactive) dlv, demo, service1, yes
# kgp - expect to see a plank being created

# (interactive squashctl) funcs
## expect to see functions

# (interactive squashctl) funcs
## expect functions to be printed

# (interactive squashctl) b main.handler
## expect something like: Breakpoint 1 set at 0x67061b for main.handler() /home/yuval/go/src/github.com/solo-io/squash/contrib/example/service1/main.go:87

# (interactive squashctl) c
## expect new line, no content


# ./sq.out # (interactive) java, demo, service2, yes, stop in io.solo.squash.service2.Service2:23
# kgp - expect to see a plank being created

### Repeat, in secure mode

# ./sq.out deploy agent
