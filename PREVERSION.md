
# Before tagging a new release

## Issues
- Ensure compatibility with vscode

## Dev tasks
- use correct namespace on CRD
- ensure the old cli args work (for ide compat)
  - esp. vscode breakpoints
- update makefile for new client/docker settings
- store DebugAttachment CRDs in the same namespace as the pods they are targeting
  - implement a CRD that lives in a known namespace that will point to all the debugging namespaces
- update the Makefile
- combine with Kube-squash
- bring vscode extension code into this repo

## Release tasks
- update the docs (no server required)
- tag a release update
- push a new image to dockerhub
- update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`

## Other
- delete this file
