# Installing Squash

To install Squash you need to install the Squash server and clients on your container orchestration platform of choice, and to install the CLI on your computer. 

### Platforms
Currently Suqash only supports Kubernets. Other platforms coming up... 
 - [Kubernetes](kubernetes.md)

### Command Line Interface (CLI)
Download the CLI binary:
- [Linux](https://github.com/solo-io/squash/releases/download/v0.2.0/squash-linux)     
- [OS X](https://github.com/solo-io/squash/releases/download/v0.2.0/squash-osx)

**For Linux**
```
curl -o squash -L https://github.com/solo-io/squash/releases/download/v0.2.0/squash-linux
```

**For Mac OS X**
```
curl -o squash -L https://github.com/solo-io/squash/releases/download/v0.2.0/squash-osx
```

Then enable execution permission:
```
chmod +x squash
```
The easiest way is to place it somewhere in your path, but it is not a must.

To make sure everything is deployed correctly, you can use `kubectl proxy` to access the Squash server pod and invoke a sample CLI command. 

1. Get the Squash server pod name by running ```kubectl get pods```. You should see something similar to this:
```
NAME                                       READY     STATUS    RESTARTS   AGE
squash-client-ds-j7fqv                     1/1       Running   0          17m
squash-client-ds-nwbkm                     1/1       Running   0          17m
squash-client-ds-zw8pp                     1/1       Running   0          17m
squash-server-rc-kwkdr                     1/1       Running   0          17m
```

2. Create a proxy to your kube api server:
```
kubectl proxy
```
You should see something like this:
```
Starting to serve on 127.0.0.1:8001
```

3. Run a list command: 
```
./squash --url=http://localhost:8001/api/v1/namespaces/default/services/squash-server:http-squash-api/proxy/api/v2 list attachments
```

The output should be like this: 
```
State |ID |tDebugger |Image |Debugger Address
```

If you have an issue with either, see the [FAQ](../faq.md) for help.