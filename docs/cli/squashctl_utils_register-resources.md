---
title: "squashctl utils register-resources"
weight: 5
---
## squashctl utils register-resources

register the custom resource definitions (CRDs) needed by squash

### Synopsis

register the custom resource definitions (CRDs) needed by squash

```
squashctl utils register-resources [flags]
```

### Options

```
  -h, --help   help for register-resources
```

### Options inherited from parent commands

```
      --config string              optional, path to squash config (defaults to ~/.squash/config.yaml)
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "mkdev")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debugger string            Debugger to use
      --json                       output json format
      --localport int              local port to use to connect to debugger (defaults to random free port)
      --machine                    machine mode input and output
      --namespace string           Namespace to debug
      --no-clean                   don't clean temporary pod when existing
      --no-guess-debugger          don't auto detect debugger to use
      --no-guess-pod               don't auto detect pod to use
      --pod string                 Pod to debug
      --process-match string       optional, if passed, Squash will try to find a process in the target container that matches (regex, case-insensitive) this string. Otherwise Squash chooses the first process.
      --squash-namespace string    the namespace where squash resources will be deployed (default: squash-debugger) (default "squash-debugger")
      --timeout int                timeout in seconds to wait for debug pod to be ready (default 300)
```

### SEE ALSO

* [squashctl utils](../squashctl_utils)	 - call various squash utils

