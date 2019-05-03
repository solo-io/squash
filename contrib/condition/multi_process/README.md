# Debug arbitrary container PIDs with Squash

At one point, Squash assumed that a user's debug target was completely described by the selection of: namespace, pod name, container. Squash assumed that the user wanted to debug the first process in the container. This is a reasonable assumption, since a popular container usage pattern is to specify one process per container. However it may be useful to run multiple processes in a single container. In order for squash to debug an arbitrary process, it needs to be told how to choose among the available processes.

# Demonstration of the properties of multi-process containers

This directory includes files needed to build and deploy a sample app as the first process in one container and the second process in a separate container.

## Build and push the containers

*This step is not needed, as the container images are already available with the values shown below. If you change these values you will need to update the manifests similarly.*

```bash
export CONTAINER_REPO_ORG = docker.io/soloio
IMAGE_TAG=v0.0.3 make all
```

## Deploy the containers

We will deploy our sample containers in their own pods:

```bash
kubectl apply -f single.yaml
kubectl apply -f multi.yaml
```

## Inspect the images

Note that the container with a single process features our app in PID 1

```bash
k exec -it squash-demo-multiprocess-base-6c746c8595-kpsvt -- /bin/s
h
/app # ls
sample_app
/app # ps
PID   USER     TIME  COMMAND
    1 root      0:00 ./sample_app
   19 root      0:00 /bin/sh
   26 root      0:00 ps
```

However, for our multi-process container, our app is not PID 1

```bash
k exec -it squash-demo-multiprocess-5fbdcd96cf-k9bzw -- /bin/sh
/app # ls
call_app.sh  sample_app
/app # ps
PID   USER     TIME  COMMAND
    1 root      0:00 {call_app.sh} /bin/sh ./call_app.sh
    7 root      0:00 ./sample_app
   20 root      0:00 /bin/sh
   27 root      0:00 ps
```

## Debug the processes with squash

You can debug the single process container without passing any flags. Squash will use the first PID by default. This works fine with our single process example.

```bash
squashctl # follow interactive prompt to choose target debug container
```

To debug a multi-process container, you need to specify a process-identifier string. Squash will look for processes whos invocation comand matches the string provided.

```bash
squashctl --process sample_app # matches with case-insensitive regex
```
