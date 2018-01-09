# Command-Line Interface

The Squash CLI wraps calls to Squash's [REST API](http://squash.solo.io) to make using Squash easy.

See intallation guide [here](install/README.md#command-line-interface-cli).

To use the Squash tool, provide the url of the squash server via the `--url` flag. If the squash server is deployed to kubernetes via the manifests included in the source code, the easiest way to do it is to run `kubectl proxy` in the background, and then and add the following url parameter: `--url=http://localhost:8001/api/v1/namespaces/squash/services/squash-server:http-squash-api/proxy/api/v2`

## Commands:
  * [`squash debug-container`](cli.md#debug-a-container)
  * [`squash debug-request`](cli.md#debug-request)
  * [`squash delete`](cli.md#delete-a-debug-config)
  * [`squash list`](cli.md#list-debug-attachments-and-requests)
  * [`squash wait`](cli.md#wwait-for-a-debug-session)

## debug a container
This command adds a debug configuration for a container (when you want to debug a known container). 
For example, to debug a go container:
```
$ squash debug-container soloio/example-service:v1.0.0 service-rc-dcjsh21  example-service dlv
Debug config id: 1427131847
```

## debug request
This command adds a debug request for a yet unknown debug attachment. The first debug attachment that has the 'match_request' set to true and matches this request will be bound to the request. 
For example, to request a debug attachment for a go service:
```
$ squash debug-request soloio/example-service:v1.0.0  dlv
Debug config id: 336122540
```

## delete a debug attachment
This command deletes a debug attachment.
Example:
```
$ squash delete 336122540
```

## list debug attachments and requests
Lists debug configs.
Example:
```
$ squash list a
State    |ID         |Debugger |Image                          |Debugger Address
attached |vnCv7CoVWe |dlv      |soloio/example-service1:v0.2.1 |squash-client-47mlm:39985
none     |jeqFghAYem |dlv      |soloio/example-service1:v0.2.1 |
```

```
$ squash list r
ID         |Debugger |Image                                      |Bound Attachment name
aYHv1cxVsz |dlv      |soloio/example-service1:v0.2.1             |vnCv7CoVWe
qOh7O8ccP5 |dlv      |soloio/example-service1:v0.2.2             |
```

## wait for a debug session
Waits for a debug session to appear in a debug config. Once the Squash client attaches a debugger, it will submit
the debug session to the Squash server. Use this command to retrieve the session debug server address from the 
Squash server.
 
Example:
```
$ squash wait vnCv7CoVWe
Debug session started! debug server is at: pod1:23421
```
