package install

import (
	"fmt"

	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/version"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	SquashRepoName  = "quay.io/solo-io"
	SquashName      = "squash"
	SquashImageName = "squash"

	ContainerPort = 1234

	volumeName = "crisock"
)

// InstallSquash creates the resources needed for Squash to run in secure mode
// If preview is set, it prints the configuration and does not apply it.
// The created resources include:
// ServiceAccount - for Squash
// ClusterRole - enabling pod creation
// ClusterRoleBinding - bind ClusterRole to Squash's ServiceAccount
// Deployment - Squash itself
func InstallSquash(cs *kubernetes.Clientset, namespace, containerRepo, containerVersion string, preview bool) error {

	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: sqOpts.SquashServiceAccountName,
		},
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: sqOpts.SquashClusterRoleName,
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
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
				Resources: []string{"debugattachments"},
				APIGroups: []string{"squash.solo.io"},
			},
			{
				Verbs:     []string{"create"},
				Resources: []string{"clusterrolebindings"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
			},
			{
				Verbs:     []string{"create"},
				Resources: []string{"clusterrole"},
				APIGroups: []string{"rbac.authorization.k8s.io"},
			},
			{
				Verbs:     []string{"create"},
				Resources: []string{"serviceaccount"},
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

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: sqOpts.SquashClusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sqOpts.SquashServiceAccountName,
				Namespace: sqOpts.SquashNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name: sqOpts.SquashClusterRoleName,
			Kind: "ClusterRole",
		},
	}

	privileged := true
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: SquashName,
			Labels: map[string]string{
				"app": SquashName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": SquashName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": SquashName,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: sqOpts.SquashServiceAccountName,
					Containers: []v1.Container{
						{
							Name:  SquashName,
							Image: fmt.Sprintf("%v/%v:%v", containerRepo, containerVersion, version.SquashImageTag),
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

	fmt.Printf("Creating service account %v\n", sqOpts.SquashServiceAccountName)
	_, err = cs.CoreV1().ServiceAccounts(namespace).Create(&sa)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Creating clusterRole %v\n", sqOpts.SquashClusterRoleName)
	_, err = cs.Rbac().ClusterRoles().Create(cr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Creating clusterRoleBinding %v\n", sqOpts.SquashClusterRoleBindingName)
	_, err = cs.Rbac().ClusterRoleBindings().Create(crb)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Creating Squash deployment")
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

	if err := cs.CoreV1().ServiceAccounts(namespace).Delete(sqOpts.SquashServiceAccountName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.Rbac().ClusterRoles().Delete(sqOpts.SquashClusterRoleName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.Rbac().ClusterRoleBindings().Delete(sqOpts.SquashClusterRoleBindingName, delOp); err != nil {
		fmt.Println(err)
	}

	if err := cs.AppsV1().Deployments(namespace).Delete(SquashName, delOp); err != nil {
		fmt.Println(err)
	}
}
