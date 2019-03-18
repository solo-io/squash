---
title: "squashctl completion"
weight: 5
---
## squashctl completion

generate auto completion for your shell

### Synopsis


	Output shell completion code for the specified shell (bash or zsh).
	The shell code must be evaluated to provide interactive
	completion of squashctl commands.  This can be done by sourcing it from
	the .bash_profile.
	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

```
squashctl completion SHELL [flags]
```

### Examples

```

	# Installing bash completion on macOS using homebrew
	## If running Bash 3.2 included with macOS
	  	brew install bash-completion
	## or, if running Bash 4.1+
	    brew install bash-completion@2
	## You may need add the completion to your completion directory
	    squashctl completion bash > $(brew --prefix)/etc/bash_completion.d/squashctl
	# Installing bash completion on Linux
	## Load the squashctl completion code for bash into the current shell
	    source <(squashctl completion bash)
	## Write bash completion code to a file and source if from .bash_profile
	    squashctl completion bash > ~/.squashctl/completion.bash.inc
	    printf "
 	     # squashctl shell completion
	      source '$HOME/.squashctl/completion.bash.inc'
	      " >> $HOME/.bash_profile
	    source $HOME/.bash_profile
	# Load the squashctl completion code for zsh[1] into the current shell
	    source <(squashctl completion zsh)
	# Set the squashctl completion code for zsh[1] to autoload on startup
	    squashctl completion zsh > "${fpath[1]}/_squashctl"
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --container string           Container to debug
      --container-repo string      debug container repo to use (default "soloio")
      --container-version string   debug container version to use (default "mkdev")
      --crisock string             The path to the CRI socket (default "/var/run/dockershim.sock")
      --debugger string            Debugger to use
      --json                       output json format
      --localport int              local port to use to connect to debugger (defaults to random free port)
      --machine                    machine mode input and output
      --namespace string           Namespace to debug
      --no-clean                   don't clean temporary pod when existing
      --no-guess-debugger          don't auto detect debugger to use
      --no-guess-pod               don't auto detect pod to use
      --pod string                 Pod to debug
      --squash-namespace string    the namespace where squash resources will be deployed (default: squash-debugger) (default "squash-debugger")
      --timeout int                timeout in seconds to wait for debug pod to be ready (default 300)
```

### SEE ALSO

* [squashctl](../squashctl)	 - debug microservices with squash

