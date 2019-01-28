package kubeutils

import (
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func GetPodNamespace(clientset *kubernetes.Clientset, podName string) (string, error) {
	if podName == "" {
		return "", fmt.Errorf("no pod name specified")
	}
	namespaces, err := GetNamespaces(clientset)
	if err != nil {
		return "", err
	}
	for _, ns := range namespaces {
		pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
		if err != nil {
			return "", fmt.Errorf("list pods for namespace %v", ns)
		}
		for _, pod := range pods.Items {
			if pod.ObjectMeta.Name == podName {
				return pod.ObjectMeta.Namespace, nil
			}
		}
	}
	return "", fmt.Errorf("pod %v not found", podName)
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

func NewKubeClientset(inCluster bool) (*kubernetes.Clientset, error) {
	if inCluster {
		return NewInClusterKubeClientset()
	}
	return NewOutOfClusterKubeClientset()
}

func NewInClusterKubeClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func NewOutOfClusterKubeClientset() (*kubernetes.Clientset, error) {
	home := homeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return &kubernetes.Clientset{}, nil
	}

	return kubernetes.NewForConfig(config)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
