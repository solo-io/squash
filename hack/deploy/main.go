package main

import (
	"flag"
	"fmt"
	"strings"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/demo"
	"k8s.io/client-go/kubernetes"
)

func main() {
	fmt.Println("deploy resource")
	if err := app(); err != nil {
		panic(err)
	}
}

func app() error {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(restCfg)
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
