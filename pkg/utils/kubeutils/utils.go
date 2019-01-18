package kubeutils

import (
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// MustGetNamespaces returns a list of all namespaces in a cluster - or panics.
// If a clientset is passed, it will use that, otherwise it creates one.
// In the event of any error it will panic.
func MustGetNamespaces(clientset *kubernetes.Clientset) []string {
	if clientset == nil {
		cs, err := NewOutOfClusterKubeClientset()
		if err != nil {
			panic(err)
		}
		clientset = cs
	}
	nss, err := GetNamespaces(clientset)
	if err != nil {
		panic(err)
	}
	return nss
}

func GetNamespaces(clientset *kubernetes.Clientset) ([]string, error) {
	namespaces := []string{}
	nss, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return namespaces, err
	}
	for _, ns := range nss.Items {
		namespaces = append(namespaces, ns.ObjectMeta.Name)
	}
	return namespaces, nil
}

func NewOutOfClusterKubeClientset() (*kubernetes.Clientset, error) {
	home := homeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return &kubernetes.Clientset{}, nil
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
