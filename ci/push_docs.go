package main

import "github.com/solo-io/go-utils/docsutils"

func main() {
	spec := docsutils.DocsPRSpec{
		Repo: "squash",
		Owner: "solo-io",
		ChangelogPrefix: "squash",
		CliPrefix: "squashctl",
		ApiPaths: []string {
			"docs/v1/github.com/solo-io/squash",
		},
		Product: "squash",
	}
	docsutils.PushDocsCli(&spec)
}
