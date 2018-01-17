
# Python debugger for VSCode

Please refer to the following [document](https://code.visualstudio.com/docs/python/debugging#_remote-debugging) 
for instructions on setting up remote debugging.

Specifically, 
- Install [ptvsd library](https://pypi.org/project/ptvsd/) 
- Add code below to the startup .py file, replacing the *passphrase* and *port number* as desired.<br>

> import ptvsd<br>
> ptvsd.enable_attach("my_secret", address = ('0.0.0.0', 3000))<br>

- In VSCode, set `"vs-squash.pythonSecret"` to the correct passphrase (`"my_secret"` in this case) in the workspace settings 
file (*select "Preferences", "Settings" and click on the "Workspace Settings" tab*)

