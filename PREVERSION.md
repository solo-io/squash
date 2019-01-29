
# Before tagging a new release

## Dev tasks
- [ ] (P0) update makefile for new client/docker settings
- [x] combine with Kube-squash
- [ ] (P1) bring vscode extension code into this repo
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
- [ ] (P0) interactive input for RBAC mode
- [ ] (P0) decide what to call RBAC/Agent
- [ ] (P0) choose container in interactive mode
- [ ] (P1, testing, docs) python support updates
- [ ] (P1, testing, docs) java support updates
- [ ] (P1, testing, docs) nodejs support updates
- [ ] (P1, future) simplify DebugAttachment.State options
## e2e tests
- [x] rbac mode
- [ ] (P1) non rbac mode

## Release tasks
- [ ] (P0) update the docs
  - [ ] (P0) no server required
  - [x] new cli flags
  - [ ] (P1) caveats for each debugger/ide
- [ ] (P0) (small) tag a release update
- [ ] (P0) (small) push a new image to dockerhub
- [ ] (P0) (small) update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`

## Other
- delete this file


# USAGE (dev)

# (in a new terminal) run the squash cli or use one of the extensions
squash debug-container soloio/example-service2:v0.2.2 example-service2-bfdcd4cf-r4cww  example-service2 --namespace squash-debugger dlv --json

# you will see your debug port printed out in the watch's terminal

dlv connect --headless localhost:32843
