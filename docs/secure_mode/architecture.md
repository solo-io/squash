
# Squash Secure Mode Architecture

## Overview

- User creates a DebugAttachment resource through `squashctl` or an IDE, which describes their debug intentions.
- Squash runs in a pod in the `squash-debugger` namespace and watches for DebugAttachment resources.
- When Squash recieves a new debug intent it creates a Plank pod in the `squash-debugger` namespace.
- The Plank pod coordinates the debugger connection in the same manner as default mode.

## Features

- Users do not need to have permission to create Plank pods in order to debug their microservices.
- Since the Squash and Plank pods are in a namespace different from the target pod, admins can configure RBAC to prevent users from exec-ing into these pods. This avoids the possibility of users exploiting the privileges of the debugging pods.
- Since a user must be able to create a DebugAttachment resource in a given namespace in order to debug a Pod in that namespace, Admins can use conventional RBAC policies to define debug permissions.
