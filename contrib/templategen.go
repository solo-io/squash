package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

type Info struct {
	Repo, Version string
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	templatetxt, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	tmpl := template.Must(template.New("versioned").Parse(string(templatetxt)))
	i := Info{
		Repo:    os.Getenv("SQUASH_REPO"),
		Version: os.Getenv("SQUASH_VERSION"),
	}
	tmpl.Execute(os.Stdout, i)

}
