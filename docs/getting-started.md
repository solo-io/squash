# Debugging your first microservice

You can debug your application from the IDE or via the CLI.

## IDEs
* [Visual Studio Code](https://github.com/solo-io/squash-vscode/blob/master/docs/example-app-kubernetes.md)


## Command Line Interface 


### Prerequisites
- A kubernetes cluster with [kubectl configured](https://kubernetes.io/docs/tasks/tools/install-kubectl/#configure-kubectl).
- Go, and DLV go debugger installed
- Squash server, client and command line binary [installed](install/README.md).
- Docker repository that you can push images to, and that kubernetes can access (docker hub for example)

### Verify
- Kubectl port-forward functionality works.
- You have access to the squash server - use `$ squash --url=http://SQUASH-SERVER-ADDRESS/api/v2 list attachments` to test that it is working properly.

If you have an issue with either, see the [FAQ](faq.md) for help.

### Build
In your favorite text editor, create a new `main.go` file. Here's the one we will be using in this tutorial:
```
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Calculator struct {
	Op1, Op2 int
	IsAdd    bool
}

func main() {
	http.HandleFunc("/calculate", calchandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func calchandler(w http.ResponseWriter, r *http.Request) {
	var req Calculator
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	isadd := req.IsAdd
	op1 := req.Op1
	op2 := req.Op2

	if isadd {
		fmt.Fprintf(w, "%d", op1-op2)
	} else {
		fmt.Fprintf(w, "%d", op1+op2)
	}
}
```

### Build a docker container
In the same folder as `main.go` add a `Dockerfile`:
```
FROM alpine
COPY microservice /microservice
ENTRYPOINT ["/microservice"]

EXPOSE 8080
```

To build everything conviently, you can add a `Makefile` (replace  <YOUR REPO HERE> with the appropreate value):
```
microservice:
	GOOS=linux CGO_ENABLED=0 go build -gcflags "-N -l" -o microservice
	docker build -t <YOUR REPO HERE>/microservice:0.1 .
dist:
	docker push <YOUR REPO HERE>/microservice:0.1
```
CGo is disabled as it is not compatible with the alpine image. The gcflags part adds more debug information for the debugger.

Over all your directory should have three files so far:
 - Dockerfile
 - main.go
 - Makefile

Finally, execute
```
$ make microservice && make dist
```
to build and deploy the microservice.

## Deploy the microservice to kubernetes.

Create a manifest for kubernetes named `microservice.yml`
```
apiVersion: v1
kind: ReplicationController
metadata:
  name: example-microservice-rc
spec:
  replicas: 1
  selector:
    app: example-microservice
  template:
    metadata:
      labels:
        app: example-microservice
    spec:
      containers:
      - name: example-microservice
        image: <YOUR REPO HERE>/microservice:0.1
        ports:
        - containerPort: 8080
          protocol: TCP
---
kind: Service
apiVersion: v1
metadata:
  name: example-microservice-svc
spec:
  selector:
    app: example-microservice
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

and deploy it to kubernets:
```
$ kubectl create -f microservice.yml
```


### Debug

Use `kubectl` to find the name of the pod you want to deubg. Then, Use the squash command line to request an attachment to a container.
```
$ squash --url=http://SQUASH-SERVER-ADDRESS/api/v2 debug-container <YOUR REPO HERE>/microservice:0.1 example-microservice-rc-n9x2r example-microservice dlv
```

Take the ID parameter from the response:

```
Debug config id: 1427131847
```

Then wait for the debugger to attach. Note - this command has a default timeout of 1 second, which may not be enough. 
If it times out, just run it again. If nothing happens after a minute or so, please [contact us](faq.md#contact).
```
$ squash --url=http://SQUASH-SERVER-ADDRESS/api/v2 wait 1427131847
Debug session started! debug server is at: squash-client-56v2q.squash:33275
```

This address is a pod.namespace:port. To easly and securly access the debugger server's port, you can use kubectl port-forward:
```
$ kubectl port-forward --namespace squash squash-client-56v2q 33275
Forwarding from 127.0.0.1:33275 -> 33275
Forwarding from [::1]:33275 -> 33275
```
Leave it running in the background. Then just attach (your ports may vary)
```
$ dlv connect localhost:33275
```

If kubectl port-forward doesnt work for you or you can't access the sqash server from your laptop, check out our [FAQ](faq.md) page.
