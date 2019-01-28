
# Before tagging a new release

## Dev tasks
- [ ] update makefile for new client/docker settings
- [ ] update the Makefile
- [x] combine with Kube-squash
- [ ] bring vscode extension code into this repo
- [x] make it easier to deploy sample apps
- [x] deployment manifest for go w/ java sample microservice

# Zoom
## namespace update
- [x] store squash in SquashCentralNamespace
- [x] allow SquashCentralNamespace to be configured as a flag
- [x] watch all namespaces for DebugAttachments
- [x] store DebugAttachment CRDs in the same namespace as the pods they are targeting
- [x] deploy sample apps from a cli tool
  - [x] go & go, same namespace
  - [x] go & go, diff namespace
  - [x] go & java, same namespace
  - [x] go & java, diff namespace
## helpers
- [x] move watch util to hack/monitor
- [x] use an event loop in hack/monitor
## remove daemonset
- [x] use pod instead of daemonset
## enhancements
- [x] squash lite supports multiple connections (use localport flag)
## cli
- [x] distinguish between lite/agent mode
- [x] use cobra
- [x] auto gen docs
## tutorial
- [x] install sample apps from command line
## backlog
- [x] allow in/out-of cluster to be configured as a flag - will not do
- [x] allow installation of squash agent from cli
## outstanding
- [ ] interactive input for RBAC mode
- [ ] decide what to call RBAC/Agent
- [ ] choose container in interactive mode
- [ ] python support updates
- [ ] java support updates
- [ ] nodejs support updates
- [ ] simplify DebugAttachment.State options
## e2e tests
- [ ] rbac mode
- [ ] non rbac mode

## Release tasks
- [ ] update the docs
  - [ ] no server required
  - [x] new cli flags
  - [ ] caveats for each debugger/ide
- [ ] tag a release update
- [ ] push a new image to dockerhub
- [ ] update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`

## Other
- delete this file


# USAGE (dev)

# (in a new terminal) run the squash cli or use one of the extensions
squash debug-container soloio/example-service2:v0.2.2 example-service2-bfdcd4cf-r4cww  example-service2 --namespace squash-debugger dlv --json

# you will see your debug port printed out in the watch's terminal

dlv connect --headless localhost:32843
