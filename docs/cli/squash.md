## squash

debug microservices with squash

### Synopsis

debug microservices with squash
	Find more information at https://solo.io

```
squash [flags]
```

### Options

```
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "v0.1.9")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debug-server               start a debug server instead of an interactive session
      --debugger string            Debugger to use (default "dlv")
  -h, --help                       help for squash
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

* [squash debug-container](squash_debug-container.md)	 - debug-container adds a container type debug config
* [squash debug-request](squash_debug-request.md)	 - debug-request adds a debug request.
* [squash deploy](squash_deploy.md)	 - deploy the squash agent or a demo microservice
* [squash list](squash_list.md)	 - lists debug requests or attachments
* [squash wait](squash_wait.md)	 - wait for a debug config to have a debug server url appear

