# Installing Squash

To install Squash you need to install the Squash server and clients on your container orchestration platform of choice, and to install the CLI on your computer. 

### Platforms
Currently Suqash only supports Kubernets. Other platforms coming up... 
 - [Kubernetes](kubernetes.md)

### Command Line Interface (CLI)
Download the CLI binary:
- [Linux](https://github.com/solo-io/squash/releases/download/v0.1.0/squash-linux)     
- [OS X](https://github.com/solo-io/squash/releases/download/v0.1.0/squash-osx)

**For Linux**
```
curl -o squash -L https://github.com/solo-io/squash/releases/download/v0.1.0/squash-linux
```

**For Mac OS X**
```
curl -o squash -L https://github.com/solo-io/squash/releases/download/v0.1.0/squash-osx
```

Then enable execution permission:
```
chmod +x squash
```
The easiest way is to place it somewhere in your path, but it is not a must.

To make sure everything is deployed correctly, you can create a port forward to the Squash server pod and invoke a sample CLI command. 

1. Get the Squash server pod name by running ```kubectl get pods```. You should see something similar to this:
```
NAME                                       READY     STATUS    RESTARTS   AGE
squash-client-ds-j7fqv                     1/1       Running   0          17m
squash-client-ds-nwbkm                     1/1       Running   0          17m
squash-client-ds-zw8pp                     1/1       Running   0          17m
squash-server-rc-kwkdr                     1/1       Running   0          17m
```

2. Create a port forward from your local machine (```http://localhost:8080```) to the ```squash-server-rc-kwkdr``` pod (note that the pod name on your machine will be different):
```
kubectl port-forward squash-server-rc-kwkdr 8080:8080 &
```
You should see something like this:
```
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

3. Run a list command: 
```
./squash --url=http://localhost:8080/api/v1 list
```

The output should be like this: 
```
Active |ID |Attachment.Name |Attachment.Type |Debugger |Image |Immediately
```

If you have an issue with either, see the [FAQ](faq.md) for help.