
# Command line
- `squashctl`
  - follow interactive prompt to choose a debugger (options include `dlv`, `java`, and `java-port`)
  - follow interactive prompt to choose a namespace and pod to debug
    - squash chooses the first process by default
  - confirm action
    - If you chose a debugger with an interactive command line interface, that interface will open in your terminal.
    - If you chose a non-interactive debugger, such as `java-port`, instrutions for connecting your debugger to the debug process will print to the screen.
- Optional quick start demo microservice:
  - For a quick start using squash to debug microservices, deploy one of the demo microservices with `squashctl deploy demo` - choose `go-go` or `go-java` apps.

# Default Mode vs. Secure Mode
- Secure mode behaves the same as default mode. Behind the scenes, Squash creates debug instances on your behalf, under the constraints your RBAC configuration.
  - This is useful if you work in a shared cluster. Rather than letting any user open a debugger on any process, you can restrict debug activities with RBAC policies.
  - This can help you avoid accidentally Squashing (or getting Squashed by) your teammates!
- To use Squash in secure mode, there are two configuration steps:
  - specify `secure_mode=true` in your Squash config file `~/.squash/config.yaml`
  - Deploy Squash to your cluster.
    - Typically your cluster admin will do this since the usefullness of secure mode comes from the associated RBAC configuration.
    - To get started quickly in secure_mode, you can also use `squashctl` to deploy Squash to your cluster with `squashctl deploy squash`
    - If you are the only user in your cluster, you should consider using the default mode since there is no risk of accidentally interferring with your teammates' activities.
