
# Before tagging a new release

## Dev tasks
- update makefile for new client/docker settings
- update the Makefile
- combine with Kube-squash
- bring vscode extension code into this repo
- make it easier to deploy sample apps
- deployment manifest for go w/ java sample microservice

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
- [ ] distinguish between lite/agent mode
- [ ] use cobra
- [ ] auto gen docs
## tutorial
- [ ] install sample apps from command line
## backlog
- [ ] allow in/out-of cluster to be configured as a flag
- [ ] allow installation of squash agent from cli

## Release tasks
- update the docs
  - no server required
  - new cli flags
  - caveats for each debugger/ide
- tag a release update
- push a new image to dockerhub
- update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`

## Other
- delete this file


# USAGE - until squash updated

# edit the active namespace
# change SquashCentralNamespace to "gloo-system" - or wherever the container you want to debug lives
vi pkg/options/default.go
	SquashCentralNamespace = "gloo-system"
    
# prevent the helper script from erroring
vi test/devutil/setup.go
comment out the "CreateNs()" block on line 63 (will error because your ns already exists)

# build the agent image
make tmpclient

# build the cli, put it in your path as "squash"
cd cmd/squash-client
go build -o squash main.go

# copy the agent to your cluster
cd test/dev
go run main.go --init

# start a watcher on the namespace
cd test/dev/watch
go run main.go --init

# (in a new terminal) run the squash cli or use one of the extensions
squash debug-container soloio/example-service2:v0.2.2 example-service2-bfdcd4cf-r4cww  example-service2 --namespace squash-debugger dlv --json

# you will see your debug port printed out in the watch's terminal

dlv connect --headless localhost:32843
