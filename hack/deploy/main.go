package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/solo-io/squash/pkg/demo"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"k8s.io/client-go/kubernetes"
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

	demoIds := []string{"go-go", "go-java"}
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
	case "go-go":
		fmt.Println("using go-go")
		return deployGoGo(cs, namespace, namespace2)
	case "go-java":
		fmt.Println("using go-java")
		return deployGoJava(cs, namespace, namespace2)
	default:
		return fmt.Errorf("Please choose a valid demo option: %v", strings.Join(demoIds, ", "))
	}

	return nil
}

func deployGoGo(cs *kubernetes.Clientset, namespace, namespace2 string) error {
	app1Name := "example-service1"
	template1Name := "soloio/example-service1:v0.2.2"

	app2Name := "example-service2"
	template2Name := "soloio/example-service2:v0.2.2"

	containerPort := 8080

	if err := demo.DeployTemplate(cs, namespace, app1Name, template1Name, containerPort); err != nil {
		return err
	}
	if err := demo.DeployTemplate(cs, namespace2, app2Name, template2Name, containerPort); err != nil {
		return err
	}
	return nil
}

func deployGoJava(cs *kubernetes.Clientset, namespace, namespace2 string) error {

	app1Name := "example-service1"
	template1Name := "soloio/example-service1:v0.2.2"

	app2Name := "example-service2-java"
	template2Name := "soloio/example-service2-java:v0.2.2"

	containerPort := 8080

	if err := demo.DeployTemplate(cs, namespace, app1Name, template1Name, containerPort); err != nil {
		return err
	}
	if err := demo.DeployTemplate(cs, namespace2, app2Name, template2Name, containerPort); err != nil {
		return err
	}
	return nil
}
