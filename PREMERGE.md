
# Before merging
## Dev-related
- use correct namespace on CRD
- ensure the old cli args work (for ide compat)
- update makefile for new client/docker settings
## Release-related
- update the docs (no server required)
- tag a release update
- push a new image to dockerhub
- update the tag in `github.com/solo-io/squash/contrib/kubernetes/squash-client.yml`
## Other
- delete this file

# After merging
## Future work
- store DebugAttachment CRDs in the same namespace as the pods they are targeting
-- implement a CRD that lives in a known namespace that will point to all the debugging namespaces
