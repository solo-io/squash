## squash debug-container

debug-container adds a container type debug config

### Synopsis

debug-container adds a container type debug config

```
squash debug-container image pod container [debugger] [flags]
```

### Options

```
  -h, --help                 help for debug-container
  -p, --processName string   Process name to debug (defaults to the first running process)
```

### Options inherited from parent commands

```
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "v0.1.9")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debug-server               start a debug server instead of an interactive session
      --debugger string            Debugger to use (default "dlv")
      --json                       output json format
      --lite                       run in lite mode (default) (default true)
      --localport int              port to use to connect to debugger (defaults to 1235)
      --machine                    machine mode input and output
      --namespace string           Namespace to debug
      --no-clean                   don't clean temporar pod when existing
      --no-detect-pod              don't auto settigns based on skaffold configuration present in current folder
      --no-guess-debugger          don't auto detect debugger to use
      --no-guess-pod               don't auto detect pod to use
      --pod string                 Pod to debug
      --timeout int                timeout in seconds to wait for debug pod to be ready (default 300)
```

### SEE ALSO

* [squash](squash.md)	 - debug microservices with squash

