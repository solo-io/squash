## squashctl deploy demo

deploy a demo microservice

### Synopsis

deploy a demo microservice

```
squashctl deploy demo [flags]
```

### Options

```
      --demo-id string           which sample microservice to deploy. Options: go-go, go-java
      --demo-namespace1 string   namespace in which to install the sample app
      --demo-namespace2 string   (optional) ns for second app - defaults to 'namespace' flag's value
  -h, --help                     help for demo
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

* [squashctl deploy](squashctl_deploy.md)	 - deploy squash or a demo microservice

