# System components
- Squash consists of three distinct processes
  - Local interface: "squashctl" (direct or via IDE extension)
  - Debugger session manager: "Plank" pod - an in-cluster pod spawned on demand for managing a particular debug session
  - RBAC expression process: "Squash" pod - (secure-mode only) an in-cluster pod that spawns Plank pods on the user's behalf according to their RBAC permissions. Typically configured by system admin


# Flow
- user declares debug intent in the command line [or in an IDE prompt]
  - debugger
  - pod namespace
  - pod name
  - container name
  - process OR matcher (both ignored right now)
- [TODO] squashctl checks that the squash-plank service account has been created
  - if not, it creates the service account and required cluster roles
- squashctl creates a crd with the debug intent
  - crd fields that are populated:
    - debug intent (fully populated: debugger, pod namespace, pod name, container name, process identifier)
    - local port
- squashctl spawns a plank [in secure mode, Squash spawns a plank]
  - plank environment variables tell it where to find the CRD
    - CRD_NAME
    - CRD_NAMESPACE
- squashctl waits for crd.plankReady=true
- plank reads crd and takes the action required for the given debugger
  - MAY:
    - start a remote debugger
  - MUST:
    - add the following information to the crd:
      - plank port
      - target port
      - plankReady = true
- squashclt port-forwards
  - port forward spec is debugger specific
- squashctl attaches local debugger
  - details are debugger specific
- squashclt waits for debug session to end
- user interacts with debugger and eventually closes it
- squashctl terminates pod (not implemented explicitly as this currently happens upon ending the debug session - may want to add a check w/ explicit delete in the future in order to ensure the old pod is removed)
- [TODO] squashctl deletes the old debugattachment crd

# Improved API outline

## API Needs

Squash requires the following information:

- way to identify plank pod
  - name
  - namespace
- way to identify target pod
  - name
  - namespace
- ports list
  - local
  - plank
  - target


## Description of upcoming API
- Intent
  - debugger
  - pod namespace
  - pod name
  - container name
  - process OR matcher (both ignored right now)
- State (for now, leave as currently exists)
- Plank information
  - pod namespace
  - pod name
  - readyForConnect
- port information
  - local port
  - plank port
  - target port


# Dev workflow notes

## setup a watcher to inspect the debug resources
```
cd test/dev/watcher
go run main
```

## initialize some sample apps and the squash client
```
cd test/dev
go run main --init # to load sample apps and squash client
go run main --att # make an attachment
go run main --clean # remove resources

# whenever you make changes to the squash client (after rebuilding)
go run main --init && go run main --clean
```

## run the e2e tests
```
cd test/e2e
export WAIT_ON_FAIL=1 # if you want better failure debugging
ginkgo -r
```

### run e2e on specific namespaces
```
go run hack/monitor/main.go -namespaces stest-1,stest-2,stest-3,stest-4,stest-5,stest-6
SERIALIZE_NAMESPACES=1 ginkgo -r
```


# Extensions
## Visual Studio Code
- install vsce
```bash
npm install -g vsce
```
- run `publish` from extension's root dir
```bash
vsce publish -p $VSCODE_TOKEN
```

# Debugger notes

## Java
- use `jdb` to attach
```bash
jdb -attach localhost:<port> -sourcepath ~/path/to/src/main/java/
```
## Go
- use 'dlv' to attach
```bash
dlv connect localhost:<port>
```
- how to specify source path
  - init file TODO(mitchdraft)


