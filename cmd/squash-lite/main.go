package main

import (
	"fmt"

	"github.com/solo-io/squash/pkg/lite/kube"
)

func main() {
	err := kube.StartDebugContainer()
	if err != nil {
		fmt.Println(err)
	}
}
