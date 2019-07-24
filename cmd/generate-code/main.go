package main

import (
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
)

func main() {
	log.Printf("Starting generate...")
	opts := cmd.GenerateOptions{
		CompileProtos: true,
		RelativeRoot:  ".",
		GenDocs: &cmd.DocsOptions{
			Output: options.Hugo,
		},
		SkipGenMocks: true,
	}
	if err := cmd.Generate(opts); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
