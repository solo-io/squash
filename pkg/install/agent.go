package install

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	AgentRepoName  = "soloio"
	AgentName      = "squash-agent"
	AgentImageName = "squash-agent"
	AgentImageTag  = "dev"

	ContainerPort = 1234

	volumeName = "crisock"

	DefaultNamespace = "squash-debugger"
)

func InstallAgent(cs *kubernetes.Clientset, namespace string) error {

	privileged := true

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: AgentName,
			Labels: map[string]string{
				"app": AgentName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": AgentName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": AgentName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  AgentName,
							Image: fmt.Sprintf("%v/%v:%v", AgentRepoName, AgentImageName, AgentImageTag),
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      volumeName,
									MountPath: "/var/run/cri.sock",
								},
							},
							SecurityContext: &v1.SecurityContext{
								Privileged: &privileged,
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									Protocol:      v1.ProtocolTCP,
									ContainerPort: int32(ContainerPort),
								},
							},
							Env: []v1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name:  "HOST_ADDR",
									Value: "$(POD_NAME).$(POD_NAMESPACE)",
								},
								{
									Name: "NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: volumeName,
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/var/run/dockershim.sock",
								},
							},
						},
					},
				},
			},
		},
	}

	// create the namespace
	fmt.Printf("Creating namespace %v\n", namespace)
	_, err := cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})
	if err != nil {
		fmt.Println(err)
	}

	crbName := fmt.Sprintf("squash-sa-cluster-admin-%v", namespace)
	fmt.Printf("Creating clusterRoleBinding %v\n", crbName)
	_, err = cs.Rbac().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: crbName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				// TODO(mitchdraft) create specific service account for squash
				Name:      "default",
				Namespace: namespace,
			},
		},
		// TODO(mitchdraft) prune these permissions
		RoleRef: rbacv1.RoleRef{
			Name: "cluster-admin",
			Kind: "ClusterRole",
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Creating squash agent deployment")
	_, err = cs.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return err
	}
	return nil
}
