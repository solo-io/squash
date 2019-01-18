package demo

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func DeployTemplate(cs *kubernetes.Clientset, namespace, appName, templateName string, containerPort int) error {

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
						},
					},
				},
			},
		},
	}

	createdDeployment, err := cs.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return err
	}
	fmt.Println(createdDeployment)

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
	createdService, err := cs.CoreV1().Services(namespace).Create(service)
	if err != nil {
		return err
	}
	fmt.Println(createdService)

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
