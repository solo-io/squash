package kubeutils

import (
	"fmt"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MustGetNamespaces returns a list of all namespaces in a cluster - or panics.
// If a clientset is passed, it will use that, otherwise it creates one.
// In the event of any error it will panic.
func MustGetNamespaces(clientset *kubernetes.Clientset) []string {
	if clientset == nil {
		restCfg, err := gokubeutils.GetConfig("", "")
		if err != nil {
			panic(err)
		}
		cs, err := kubernetes.NewForConfig(restCfg)
		if err != nil {
			panic(err)
		}
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
