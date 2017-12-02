# Command-Line Interface

The Squash CLI wraps calls to Squash's [REST API](http://squash.solo.io) to make using Squash easy.

See intallation guide [here](install/README.md#command-line-interface-cli).

To use the Squash tool, provide the url of the squash server via the `--url` flag. If the squash server is deployed to kubernetes via the manifests included in the source code, the easiest way to do it is to run `kubectl proxy` in the background, and then and add the following url parameter: `--url=http://localhost:8001/api/v1/namespaces/default/services/squash-server-service/proxy/api/v1`

## Commands:
  * [`squash debug-container`](cli.md#debug-a-container)
  * [`squash debug-service`](cli.md#debug-a-service)
  * [`squash delete`](cli.md#delete-a-debug-config)
  * [`squash list`](cli.md#list-debug-configs)
  * [`squash wait`](cli.md#wwait-for-a-debug-session)

## debug a container
This command adds a debug configuration for a container (when you want to debug a known container). 
For example, to debug a go container:
```
$ squash debug-container soloio/example-service:v1.0.0 service-rc-dcjsh21  example-service dlv
Debug config id: 1427131847
```

## debug a service
This command adds a debug configuration for a service. A debug session will be created for the first contaienr 
of the service that generates a debug event (crash or breakpoint). 
For example, to debug a go service:
```
$ squash debug-container service-name soloio/example-service:v1.0.0  dlv --breakpoint main.go:80
Debug config id: 336122540
```

## delete a debug config
This command deletes a debug configuration.
Example:
```
$ squash delete 336122540
```

## list debug configs
Lists debug configs.
Example:
```
$ squash list
Active |ID         |Attachment.Name |Attachment.Type |Debugger |Image  |Immediately
true   |336122540  |pod1:container1 |service         |dlv      |image1 |false
true   |1298498081 |pod4:container1 |container       |dlv      |image1 |true
false  |2019727887 |pod3:container2 |container       |dlv      |image2 |true
false  |1427131847 |pod2:container2 |container       |dlv      |image2 |true

```

## wait for a debug session
Waits for a debug session to appear in a debug config. Once the Squash client attaches a debugger, it will submit
the debug session to the Squash server. Use this command to retrieve the session debug server address from the 
Squash server.
 
Example:
```
$ squash wait 336122540
Debug session started! debug server is at: pod1:23421
```
