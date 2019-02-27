package main

import (
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func main() {
	log.Printf("Starting generate...")
	docsOpts := cmd.DocsOptions{
		Output: options.Hugo,
	}
	if err := cmd.Run(".", true, &docsOpts, nil, []string{}); err != nil {
		log.Fatalf("generate failed!: %v", err)
	}
}
