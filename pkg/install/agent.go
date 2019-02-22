package install

import (
	"fmt"

	"github.com/solo-io/squash/pkg/version"
	"gopkg.in/yaml.v2"
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

	ContainerPort = 1234

	volumeName = "crisock"

	DefaultNamespace = "squash-debugger"

	squashServiceAccountName     = "squash"
	squashClusterRoleName        = "squash-cr-pods"
	squashClusterRoleBindingName = "squash-crb-pods"
)

// InstallAgent creates the resources needed for Squash to run in secure mode
// If preview is set, it prints the configuration and does not apply it.
// The created resources include:
// ServiceAccount - for Squash
// ClusterRole - enabling pod creation
// ClusterRoleBinding - bind ClusterRole to Squash's ServiceAccount
// Deployment - Squash itself
func InstallAgent(cs *kubernetes.Clientset, namespace string, preview bool) error {

	// Squash ServiceAccount
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: squashServiceAccountName,
		},
	}
	fmt.Println(sa)

	// Squash ClusterRole
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: squashClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
				Resources: []string{"pods"},
				APIGroups: []string{""},
			},
			{
				Verbs:     []string{"list"},
				Resources: []string{"namespaces"},
				APIGroups: []string{""},
			},
			{
				// TODO remove the register permission when solo-kit is updated
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete", "register"},
				Resources: []string{"customresourcedefinitions"},
				APIGroups: []string{"apiextensions.k8s.io"},
			},
		},
	}
	fmt.Println(cr)

	// Squash ClusterRoleBinding
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      squashClusterRoleBindingName,
			Namespace: DefaultNamespace,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      squashServiceAccountName,
				Namespace: DefaultNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name: squashClusterRoleName,
			Kind: "ClusterRole",
		},
	}

	// Squash Deployment
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
					ServiceAccountName: squashServiceAccountName,
					Containers: []v1.Container{
						{
							Name:  AgentName,
							Image: fmt.Sprintf("%v/%v:%v", AgentRepoName, AgentImageName, version.AgentImageTag),
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

	if preview {
		// TODO - also include permissions etc
		// TODO - use k8s printer to avoid printing null values
		// TODO - produce valid yaml for `kubectl apply -f`
		if err := simplePrinter(crb); err != nil {
			return err
		}
		if err := simplePrinter(deployment); err != nil {
			return err
		}
		return nil
	}

	// create the resources
	fmt.Printf("Creating namespace %v\n", namespace)
	_, err := cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Creating service account %v\n", squashServiceAccountName)
	_, err = cs.CoreV1().ServiceAccounts(namespace).Create(&sa)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Creating clusterRole %v\n", squashClusterRoleName)
	_, err = cs.Rbac().ClusterRoles().Create(cr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Creating clusterRoleBinding %v\n", squashClusterRoleBindingName)
	_, err = cs.Rbac().ClusterRoleBindings().Create(crb)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Creating squash agent deployment")
	_, err = cs.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		cleanupDeployment(cs, namespace)
		return err
	}
	return nil
}

func simplePrinter(val interface{}) error {
	yml, err := yaml.Marshal(val)
	if err != nil {
		return err
	}
	fmt.Println(string(yml))
	return nil
}

func cleanupDeployment(cs *kubernetes.Clientset, namespace string) {
	delOp := &metav1.DeleteOptions{}

	if err := cs.CoreV1().ServiceAccounts(namespace).Delete(squashServiceAccountName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.Rbac().ClusterRoles().Delete(squashClusterRoleName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.Rbac().ClusterRoleBindings().Delete(squashClusterRoleBindingName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.AppsV1().Deployments(namespace).Delete(AgentName, delOp); err != nil {
		fmt.Println(err)
	}
}
