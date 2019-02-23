<h1 align="center">
    <img src="docs/squash.svg" alt="squash" width="200" height="248">
  <br>
  Debugger for microservices
</h1>


<h4 align="center">Debug your microservices application running on container orchestration from IDE.</h4>
<BR>


Debugging microservices applications is a difficult task. The state of the application is spread across multi microservices and it is hard to get the holistic view of the state of the application. Currently debugging of microservices is assisted by openTracing, which helps in tracing of a transaction or workflow for post-mortem analysis, and service mesh like Istio which monitor the network to identify latency problems. These tools however, do not allow to monitor and interfere with the application during run time. 

In contrast, "traditional" debuggers of monolithic application provide devs with powerful features like setting breakpoints in their codes, following values of variables on the fly, stepping through the code, and changing these variables during run time. 

Squash brings the power of modern popular debuggers to developers of microservices apps that run on container orchestration platforms. Squash bridges between the orchestration platform (without changing it) and IDE. Users are free to choose which containers, pods, services or images they are interested in debugging, and are allowed to set breakpoints in their codes, follow values of their variables on the fly, step through the code while jumping between microservices, and change these values during run time. 

Squash is built to be easily extensible, allowing – and encouraging – adding support for more platforms, debuggers and IDEs.


To learn more about the motivation behind project squash, read our blog [post](https://medium.com/solo-io/squash-microservices-debugger-5023e27533de) or [watch](https://www.infoq.com/presentations/squash-microservices-container) session ([slides](https://www.slideshare.net/IditLevine/debugging-microservices-qcon-2017)). We also encourage you to read squash technical overview [blog](https://medium.com/solo-io/technical-introduction-to-squash-399e0c0c54b).

To stay up-to-date with Squash, follow us [@GetSoloIO](https://twitter.com/GetSoloIO) and join us on our [slack channel](http://slack.solo.io).


## With Squash, you can:
* Live debugging cross multi microservices
* Debug container in a pod
* Debug a service
* Set breakpoints
* Step through the code
* View and modify values of variables
* and more ...


## Demo

In the following demo we debug an application that adds two numbers. As you can see, it currently fails miserably at adding 9 to 99. The application is composed of two microservices. We  set breakpoints in both, then step through the application, while monitoring its variables. At some point we identify the problem, and test it by changing the value of the variable isadd before resuming the execution of the application.

<img src="images/squash-demo-2.gif" alt="Squash Demo" />

An annotated version of this demo can be found [here](https://youtu.be/5aNPfwVvLvA).


## Documentation
- **Installation**
  - [install squash](docs/install)
- **Getting Started**
  - [debug your microservice](docs/getting-started.md)
- **User Documentation**
  - using [IDEs to debug](docs/IDEs.md)
  - using the [command line interface](docs/cli.md)
  - [Debug your java microservices](docs/debuggers/java.md)
  - [Debug your NodeJS microservices](docs/debuggers/nodejs.md)
  - [Debug your python microservices with VSCode](docs/debuggers/python_vscode.md)
  - [Debug application using service mesh](docs/platforms/istio.md)

- **Developer Documentation**
  - how to [build squash](docs/build) from source
  - [technical overview](docs/techincal-overview.md)
  - adding [debugger](docs/debuggers.md) support
  - adding [platform](docs/platforms.md) support
  - squash's [REST API](http://squash.solo.io)

---

## Supported debuggers:
 - [gdb](https://www.gnu.org/software/gdb/)
 - [dlv](https://github.com/derekparker/delve)
 - [Java](http://docs.oracle.com/javase/7/docs/technotes/guides/jpda/jdwp-spec.html)
 - [Nodejs](https://nodejs.org/api/debugger.html)
 - [Python - ptvsd](https://code.visualstudio.com/docs/python/debugging)
 
## Supported platforms:
 - [Kubernetes](docs/platforms/kubernetes.md)
 - [Istio](docs/platforms/istio.md)
 
## Supported IDEs:
 - [VS Code](https://github.com/solo-io/squash-vscode)
 - [Intellij](https://github.com/solo-io/squash-intellij)

*We are looking for community help to add support for more debuggers and IDEs.*

## Roadmap:
**Service Mesh**
  - Squash integrates with [Envoy](https://www.envoyproxy.io). Read about the Squash HTTP filter, now part of Envoy [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/squash_filter.html). This allows Squash to open debug sessions as a request flows through a microservice. Support for Istio will be added in 2019.

**debuggers**
 - [Python - pdb](https://docs.python.org/3/library/pdb.html)

**IDEs**
  - [Eclipse](https://eclipse.org/ide/)


Squash is under active development. APIs and compatibility are subject to change. We welcomd community participation to help identify potential bugs and compatibility issues. Please open a Github issue for any problems you may experience, and join us on our [slack channel](http://slack.solo.io)

---

## Thanks

**Squash** would not be possible without the valuable open-source work of projects in the community. We would like to extend a special thank-you to [Kubernetes](https://kubernetes.io), [gdb](https://www.gnu.org/software/gdb/) and [dlv](https://github.com/derekparker/delve).
