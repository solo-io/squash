package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatalf("%v", err)
	}
}
