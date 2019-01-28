## squash agent

manage the squash agent

### Synopsis

Squash agent allows you to debug with RBAC enabled.
This may be desired for shared clusters. Squash supports RBAC through its agent
mode. In this mode, a squash agent runs in your cluster as a server. When a
user initiates a debug session with squash, the agent will create the debugging
pod, rather than the person who initiated the debug session. The agent will
only open debug session on pods in namespaces where you have CRD write access.
You can configure squash to use agent mode by setting the ENABLE_RBAC_MODE value
in your .squash config file.


### Options

```
  -h, --help   help for agent
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
      --no-guess-debugger          don't auto detect debugger to use
      --no-guess-pod               don't auto detect pod to use
      --pod string                 Pod to debug
      --timeout int                timeout in seconds to wait for debug pod to be ready (default 300)
```

### SEE ALSO

* [squash](squash.md)	 - debug microservices with squash
* [squash agent delete](squash_agent_delete.md)	 - delete squash agents by namespace
* [squash agent status](squash_agent_status.md)	 - list status of squash agent

