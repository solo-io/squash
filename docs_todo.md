

### Content to add to the docs

#### IDE
- vscode java mode requires the java debug extension https://github.com/Microsoft/vscode-java-debug
- you may need to update your JDK if you get a [java crash error](https://github.com/redhat-developer/vscode-java/issues/681)
- workspace configuration:
  - `remotePath`: filepath used for source code mapping
  - `path`: path to `squashctl` binary
```json
{
  "squashextension.remotePath": "/home/yuval/go/src/github.com/solo-io/squash/contrib/example/service2-java",
  "squashextension.path": "/Users/mitch/go/src/github.com/solo-io/squash/target/squashctl",
}
```

- how to resolve this warning:
```
[Warn] The debugger and the debuggee are running in different versions of JVMs. You could see wrong source mapping results.
Debugger JVM version: 11.0.2
Debuggee JVM version: 1.8.0_111
```
