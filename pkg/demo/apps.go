package demo

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

var (
	GoApp1Name      = "example-service1"
	GoTemplate1Name = "soloio/example-service1:v0.2.2"

	GoApp2Name      = "example-service2"
	GoTemplate2Name = "soloio/example-service2:v0.2.2"

	JavaApp2Name      = "example-service2-java"
	JavaTemplate2Name = "soloio/example-service2-java:v0.2.2"

	DemoGoGo   = "go-go"
	DemoGoJava = "go-java"
	DemoIds    = []string{DemoGoGo, DemoGoJava}
)

func DeployTemplate(cs *kubernetes.Clientset, namespace, appName, templateName, appType string, containerPort int) error {

	service2url := ""
	if appType == DemoGoJava {
		service2url = "example-service2-java"
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": appName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  appName,
							Image: templateName,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: int32(containerPort),
								},
							},
							Env: []v1.EnvVar{{
								Name:  "SERVICE2_URL",
								Value: service2url,
							}},
						},
					},
				},
			},
		},
	}

	_, err := cs.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return err
	}

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": appName,
			},
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: v1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: int32(containerPort),
					},
				},
			},
		},
	}
	_, err = cs.CoreV1().Services(namespace).Create(service)
	if err != nil {
		return err
	}
	fmt.Printf("Deployed demo app %v in namespace %v\n", appName, namespace)

	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func DeployGoGo(cs *kubernetes.Clientset, namespace, namespace2 string) error {
	fmt.Println("deploying go-go sample microservice")
	app1Name := GoApp1Name
	template1Name := GoTemplate1Name

	app2Name := GoApp2Name
	template2Name := GoTemplate2Name

	containerPort := 8080

	if err := DeployTemplate(cs, namespace, app1Name, template1Name, DemoGoGo, containerPort); err != nil {
		return err
	}
	if err := DeployTemplate(cs, namespace2, app2Name, template2Name, DemoGoGo, containerPort); err != nil {
		return err
	}
	return nil
}

func DeployGoJava(cs *kubernetes.Clientset, namespace, namespace2 string) error {
	fmt.Println("deploying go-java sample microservice")

	app1Name := GoApp1Name
	template1Name := GoTemplate1Name

	app2Name := JavaApp2Name
	template2Name := JavaTemplate2Name

	containerPort := 8080

	if err := DeployTemplate(cs, namespace, app1Name, template1Name, DemoGoJava, containerPort); err != nil {
		return err
	}
	if err := DeployTemplate(cs, namespace2, app2Name, template2Name, DemoGoJava, containerPort); err != nil {
		return err
	}
	return nil
}
