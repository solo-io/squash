
# NodeJS debugger

There are two different debugger options:
* nodejs - for microservices running Node JS version below 8 (using port 5858)
* nodejs8 - for Node JS version 8+ (using port (9229)

Additionally, name of the application ("node") often needs to be provided to squash to be able to find the correct process ID.
* From command line - use `"-p node"`.
* When using VSCode set `"vs-squash.process-name"` to `"node"` in the workspace settings file (*select "Preferences", "Settings" and click on 
the "Workspace Settings" tab*)

