package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/solo-io/squash/pkg/demo"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
)

func main() {
	fmt.Println("deploy resource")
	if err := app(); err != nil {
		panic(err)
	}
}

func app() error {
	cs, err := kubeutils.NewOutOfClusterKubeClientset()
	if err != nil {
		return err
	}

	namespace := "squash"
	namespace2 := ""

	demoId := "default"

	flag.StringVar(&namespace, "namespace", "default", "namespace in which to install the sample app")
	flag.StringVar(&namespace2, "namespace2", "", "(optional) ns for second app - defaults to 'namespace' flag's value")
	flag.StringVar(&demoId, "demo", "default", "which sample microservice to deploy. Options: go-go, go-java")
	flag.Parse()

	if namespace2 == "" {
		namespace2 = namespace
	}

	switch demoId {
	case "default":
		fallthrough
	case demo.DemoGoGo:
		fmt.Println("using go-go")
		return demo.DeployGoGo(cs, namespace, namespace2)
	case demo.DemoGoJava:
		fmt.Println("using go-java")
		return demo.DeployGoJava(cs, namespace, namespace2)
	default:
		return fmt.Errorf("Please choose a valid demo option: %v", strings.Join(demo.DemoIds, ", "))
	}

	return nil
}
