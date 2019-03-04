---
title: "squashctl"
weight: 5
---
## squashctl

debug microservices with squash

### Synopsis

Squash requires no arguments. Just run it!
It creates a privileged debug pod, starts a debugger, and then attaches to it.
If you are debugging in a shared cluster, consider using Squash (in cluster process).
(squashctl squash --help for more info)
Find more information at https://solo.io


```
squashctl [flags]
```

### Options

```
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "mkdev")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debugger string            Debugger to use
  -h, --help                       help for squashctl
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

* [squashctl completion](../squashctl_completion)	 - generate auto completion for your shell
* [squashctl deploy](../squashctl_deploy)	 - deploy squash or a demo microservice
* [squashctl squash](../squashctl_squash)	 - manage the squash
* [squashctl utils](../squashctl_utils)	 - call various squash utils

