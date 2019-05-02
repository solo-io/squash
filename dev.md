# Extensions
## Visual Studio Code
- install vsce
```bash
npm install -g vsce
```
- run `publish` from extension's root dir
```bash
vsce publish -p $VSCODE_TOKEN
```

# Debugger notes

## Java
- use `jdb` to attach
```bash
jdb -attach localhost:<port> -sourcepath ~/path/to/src/main/java/
```
## Go
- use 'dlv' to attach
```bash
dlv connect localhost:<port>
```
- how to specify source path
  - init file TODO(mitchdraft)


