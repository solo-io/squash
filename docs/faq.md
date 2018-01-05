
# Troubleshoot
## kubectl port-forward doesn't work

Sometimes kubectl port-forward may not work - as it needs a less resitricted network access.
For example, kubernetes deployed on AWS via kops, has tight security groups, and kubectl portforward will not work on a laptop outside AWS.

The solution for this is to provide it with a secure way in. Normally, ssh is to the rescue. ssh can create a SOCKS proxy for us. The problem is that kubectl doesn't support SOCKS proxy. To solve that, we use "polipo" to 'convert' the socks proxy to an
http proxy.

ssh to a node on aws and create a socks proxy:
```
ssh -N -D 12346  admin@[NODE in the cluster here]
```
'convert' the socks proxy to an http proxy:
```
docker run --rm --net=host clue/polipo proxyAddress=127.0.0.1 proxyPort=12347 socksParentProxy="localhost:12346" socksProxyType=socks5 allowedPorts=1-65535 tunnelAllowedPorts=1-65535
```

then set an http proxy for kubectl.
You can do it as an env var before starting vscode:
```
export http_proxy=localhost:12347
```
Or you can use the "vs-squash.kubectl-proxy" setting in vscode. This setting is very focused and will only apply for the kubectl port-forward call.

kubectl port-forward will now work.

## Can't access squash server
We can use kubectl to overcome this issue. Run this command to start kubectl's api proxy:
```
kubectl proxy
```

We can then use kubernetes' service proxy - the squash server url will be available as: `http://localhost:8001/api/v1/namespaces/squash/services/squash-server:http-squash-api/proxy/api/v2`.

You can use the squash client now with a url flag:
```
$ squash --url=http://localhost:8001/api/v1/namespaces/squash/services/squash-server:http-squash-api/proxy/api/v2 ...
```

Or add this setting to vs-code:
```
"vs-squash.squash-server-url": "http://localhost:8001/api/v1/namespaces/squash/services/squash-server:http-squash-api/proxy/api/v2"
```

Just note that `kubectl proxy` should remain running in the background.

# Permissions
## Why does the squash client needs to be privilged?
The daemon set needs to be priviledged to be able to debug processes.

## Why does the squash client needs to be in the host pid namespace?
It needs to be in the hosts PID namespace and order to "see" the process to debug.

## Why does the squash client needs access to the CRI socket interface?
The squash client uses the CRI interface to understand what is the process-id of the container which we want to debug.

# Contact
## What information should I include in an issue?
## How to submit patches?
Please use github's pull requests
