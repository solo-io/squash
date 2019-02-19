#! /bin/bash

eval $(minikube docker-env)
make pushlocal
