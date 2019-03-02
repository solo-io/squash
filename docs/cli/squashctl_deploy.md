---
title: "squashctl deploy"
weight: 5
---
## squashctl deploy

deploy squash or a demo microservice

### Synopsis

deploy squash or a demo microservice

### Options

```
  -h, --help   help for deploy
```

### Options inherited from parent commands

```
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "mkdev")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debug-server               [deprecated] start a debug server instead of an interactive session
      --debugger string            Debugger to use
      --json                       output json format
      --localport int              local port to use to connect to debugger (defaults to random free port)
      --machine                    machine mode input and output
      --namespace string           Namespace to debug
      --no-clean                   don't clean temporary pod when existing
      --no-guess-debugger          don't auto detect debugger to use
      --no-guess-pod               don't auto detect pod to use
      --pod string                 Pod to debug
      --squash-namespace string    the namespace where squash resourcea will be deployed (default: squash-debugger) (default "squash-debugger")
      --timeout int                timeout in seconds to wait for debug pod to be ready (default 300)
```

### SEE ALSO

* [squashctl](../squashctl)	 - debug microservices with squash
* [squashctl deploy demo](../squashctl_deploy_demo)	 - deploy a demo microservice
* [squashctl deploy squash](../squashctl_deploy_squash)	 - deploy Squash to cluster

