
# Before tagging a new release

## General
- [x] combine with Kube-squash
- [x] make it easier to deploy sample apps
- [x] deployment manifest for go w/ java sample microservice
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
- [x] (P0) interactive input for RBAC mode
- [x] (P0) decide what to call RBAC/Agent
- [x] (P0) choose container in interactive mode
- [ ] (P1, testing, docs) python support updates
- [ ] (P1, testing, docs) java support updates
- [ ] (P1, testing, docs) nodejs support updates
- [ ] (P1, future) simplify DebugAttachment.State options
## e2e tests
- [x] rbac mode
- [ ] (P1) non rbac mode
## Wrap up
- [x] squashctl: wait for pod to be created before expecting crd
- [ ] clean up artifacts: (auto) delete debug attachment crd on exit
- [ ] clean up artifacts: (prompted) delete permissions
- [ ] check for existence of permissions before creating them
- [ ] only notify when newly creating permissions
- [ ] handle error where user tries to create a second debug attachment on a single process
- [ ] use case: java debug with port-forward only - should print port info and wait for close. Can be implemented as an alternative local java debugger "java-port" for example
- [ ] security: in secure mode only spawn planks in the squash-debugger namespace
- [ ] security: add documentation suggesting that users not be allowed to exec into any pod running in the squash-debugger namespace (per Dio's suggestion)
- [ ] permissions: fix permissions on planks created in secure mode
- [x] bug: agent deletes crd before squashctl can read it's values - need to rework secure-mode crd lifecycle


## Release tasks
- [x] (P0) update makefile for new client/docker settings
- [ ] (P0) update the docs
  - [ ] (P0) no server required
  - [x] new cli flags
  - [ ] (P1) caveats for each debugger/ide
- [ ] (P0) (small) tag a release update
- [ ] (P0) (small) push a new image to dockerhub
- [ ] (P0) (small) update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`

## Port from kubesquash
- [x] (P0) there are a few kubesquash commits that were contributed recently that we should port.
- [ ] (P0) also port the CI process that releases everything automatically.
- [ ] (P0) port and merge the extensions (i.e. port the download an version pinning from the kube squash extension)

## Other
- delete this file


# USAGE (dev)

# (in a new terminal) run the squash cli or use one of the extensions
squash debug-container soloio/example-service2:v0.2.2 example-service2-bfdcd4cf-r4cww  example-service2 --namespace squash-debugger dlv --json

# you will see your debug port printed out in the watch's terminal

dlv connect --headless localhost:32843


# Detangle secure, unsecure, ide

## vs code kubesquash extension:
```
let stdout = await exec(maybeKubeEnv() + `${squahspath} ${containerRepoArg} -machine -debug-server -pod ${selectedPod.metadata.name} -namespace ${selectedPod.metadata.namespace}`);
```
### flags:
* -machine
* -debug-server
* -pod ${selectedPod.metadata.name}
* -namespace ${selectedPod.metadata.namespace}

### notes
Since machine and debug server are passed together, we probably don't need both

## vs code squash extension
```
let cmdline = `debug-container --namespace=${podnamespace} ${imgid} ${podname} ${container} ${dbgr}`
```
### flags:
* --namespace=${podnamespace}
### args (positional):
1. debug-container
2. ${imgid}
3. ${podname}
4. ${container}
5. ${dbgr} # dlv, etc


## Machine
- default false
- set true by vscode kubesquash
- gates confirmation prompt

## DebugServer
- bool, default false
- description is similar to "Machine"
- set true by vscode kubesquash
- redundant with Machine?

## InCluster
- made by mitch during refactor
- gates interactive gathering
- -replaced by Machine


# Naming
## Command line tools
- squashctl - the command line tool that the user uses to initiate debug sessions
## Pods
- squash - watches for debug session requests (via DebugAttachment CRDs) and creates and removes squash-debugger pods
- squash-debugger - a pod that runs a debugger process
## Modes of operation
- secure-mode - applies RBAC policy to debugging permissions
