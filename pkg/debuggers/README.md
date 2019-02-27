# Debugger Interfaces

- Squash manages both remote and local debuggers
- **Remote debuggers** run in the cluster
  - Used by `squash` and `plank`
  - Runs in a Linux container
- **Local debuggers** run in your local environment
  - Used by squashctl
  - Runs in for any Golang OS (linux, mac, windows)

# Adding new debuggers
- it's easy to add new debuggers, just implement the remote and local interfaces
