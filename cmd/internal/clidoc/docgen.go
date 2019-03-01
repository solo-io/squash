package clidoc

// DO NOT USE THIS PACKAGE
// This package will be moved to go-utils and thereafter deleted from this repo

import (
	"bytes"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// GenerateCliDocs is the official way to convert Solo.io's command line tools to online documentation.
// It applies the file formatting and directory placement expected by Solo's documentation conventions.
func GenerateCliDocs(app *cobra.Command) error {
	disableAutoGenTag(app)
	linkHandler := func(s string) string {
		if strings.HasSuffix(s, ".md") {
			return filepath.Join("..", s[:len(s)-3])
		}
		return s
	}
	return doc.GenMarkdownTreeCustom(app, "./docs/cli", renderFrontMatter, linkHandler)
}

// MustGenerateCliDocs is the same as GenerateCliDocs but it exits with status 1 on error
func MustGenerateCliDocs(app *cobra.Command) {
	if err := GenerateCliDocs(app); err != nil {
		log.Fatal(err)
	}
}

const frontMatter = `---
title: "{{ replace .Name "_" " " }}"
weight: 5
---
`

var funcMap = template.FuncMap{
	"title":   strings.Title,
	"replace": func(s, old, new string) string { return strings.Replace(s, old, new, -1) },
}

var frontMatterTemplate = template.Must(template.New("frontmatter").Funcs(funcMap).Parse(frontMatter))

func renderFrontMatter(filename string) string {
	_, justFilename := filepath.Split(filename)
	ext := filepath.Ext(justFilename)
	justFilename = justFilename[:len(justFilename)-len(ext)]
	info := struct {
		Name string
	}{
		Name: justFilename,
	}
	var buf bytes.Buffer
	err := frontMatterTemplate.Execute(&buf, info)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func disableAutoGenTag(c *cobra.Command) {
	c.DisableAutoGenTag = true
	for _, c := range c.Commands() {
		disableAutoGenTag(c)
	}
}
