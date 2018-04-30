#Squash Lite!

## What is it?
A squash version with just a command line client and NO server to deploy.

Right now squash lite only works for kubernetes, debugging go programs using dlv.

## Prerequisite
Just `kubectl` configured to your cluster. And obviously some go program to debug.

## How to use it?

Download the binary, and then just run it. You will be asked to input the pod you wish to debug. You will then be presented with a command line dlv prompt.

## Status
This is a very initial version of squash lite. We released it early to get community feedback.

# Future plans:

- VSCode integration
- More debuggers (python, java..)
- Better Skaffold integration (to autodetect settings)

# To Use
Grab squash lite from our releases page:

https://github.com/solo-io/squash/releases

# To Build

```
make DOCKER_REPO=your-docker-repo target/squash-lite-container-pushed
make DOCKER_REPO=your-docker-repo target/squash-lite
```