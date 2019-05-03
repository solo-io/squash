---
title: "squashctl squash"
weight: 5
---
## squashctl squash

manage the squash

### Synopsis

Squash allows you to debug with RBAC enabled.
This may be desired for shared clusters. Squash supports RBAC through its secure.
mode. In this mode, a Squash runs in your cluster as a server. When a
user initiates a debug session with Squash, Squash will create the debugging
pod, rather than the person who initiated the debug session. Squash will
only open debug session on pods in namespaces where you have CRD write access.
You can configure squash to use secure mode by setting the secure_mode value
in your .squash config file.


```
squashctl squash [flags]
```

### Options

```
  -h, --help   help for squash
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

* [squashctl](../squashctl)	 - debug microservices with squash
* [squashctl squash delete](../squashctl_squash_delete)	 - delete Squash processes from your cluster by namespace
* [squashctl squash status](../squashctl_squash_status)	 - list status of Squash process

