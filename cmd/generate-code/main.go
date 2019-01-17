package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func main() {
	relativeRoot := "./api/v1"
	compileProtos := true
	genDocs := true

	if err := cmd.Run(relativeRoot, compileProtos, genDocs, []string{}, []string{}); err != nil {
		log.Fatalf("%v", err)
	}
}
