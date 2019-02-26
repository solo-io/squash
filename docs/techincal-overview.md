# Technical Overview

## Squash Key concepts:
- Squash consists of three distinct processes
  - **Local interface: `squashctl`** (via terminal or IDE extension)
  - **Debugger session manager: "Plank" pod** - an in-cluster pod spawned on demand for managing a particular debug session
  - **RBAC expression process: "Squash" pod** - (secure-mode only) an in-cluster pod that spawns Plank pods on the user's behalf according to their RBAC permissions. Typically configured by system admin.


## Constraints
- In order to debug a program, you need to have a debugger on the same node, with the process to be debugged visible to it.
- Internal structure of the program can change every time a program is compiled.

## Solution

Given the constraints, we need to run a Plank pod on the same node as the target process. The Plank process shares the hosts PID namespace (and hence can see all processes on the node, making them available to be debugged). 

To debug a process, the user selects it through one of Squash's interactive interfaces. The interface triggers the creation of a Plank pod. The Plank pod is initiated with knowledge of the user's debug intentions.

Next, the Plank process determines which pid the user wants to debug. To do that, it obains a list of all the 
currently running pids in that container using the Container Runtime Interface (CRI) API. It does so in a CRI agnostic way, with the only assumption that "ls" is present inside the image.

Squash attaches a debugger to the target pid and exposes the debugger to the user. Depending on the language and user preferences, the debugger interface is exposed directly in the context of an IDE, in a terminal-based interactive debugger, or simply as a local port that they can attach to their preferred debug interface.
