#!/bin/bash
set -x
set -o errexit
set -o nounset
set -o pipefail

../../../../../../../k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis \
  squash:v1 \
  --go-header-file boilerplate.go.txt

