# Sample Microservices

## Overview
- this directory contains a simple "calculator" microservice consisting of two services
- service 1: display logic
  - implemented in Golang only
- service 2: computation logic
  - implemented in Golang and Java

## Go-Go
- deploy
```bash
squashctl deploy demo
# interactive prompt: choose namespaces for each service, choose go-go (or go-java)
```

## Go-Java
- deploy
```bash
squashctl deploy demo
# interactive prompt: choose namespaces for each service, choose go-java (or go-go)
```


# Go (dlv) debug tips
- set a breakpoint
```bash
b main.handler
```
- execute current line of code
```bash
s
```
- resume execution
```bash
c
```
- pause execution and create a breakpoint
```bash
<control-c> # enter dlv control mode
s # stop the target
```

# Java (jdb) debug tips
- set a breakpoint
```bash
stop in io.solo.squash.service2.Service2:23
```
- execute current line of code
```bash
next
```
- resume execution
```bash
cont
```
