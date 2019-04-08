package kubeutils

import (
	"github.com/solo-io/go-utils/errors"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// required to add support for auth plugins such as oidc, azure, etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// MustGetNamespaces returns a list of all namespaces in a cluster - or panics.
// If a clientset is passed, it will use that, otherwise it creates one.
// In the event of any error it will panic.
func MustGetNamespaces(clientset *kubernetes.Clientset) ([]string, error) {
	if clientset == nil {
		cs, err := GetKubeClient()
		if err != nil {
			return nil, err
		}

		clientset = cs
	}

	nss, err := GetNamespaces(clientset)
	if err != nil {
		return nil, err
	}
	return nss, nil
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

func GetKubeClient() (*kubernetes.Clientset, error) {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return &kubernetes.Clientset{}, errors.Wrapf(err, "no Kubernetes context config found; please double check your Kubernetes environment")
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return &kubernetes.Clientset{}, errors.Wrapf(err, "error connecting to current Kubernetes Context Host %s; please double check your Kubernetes environment", restCfg.Host)
	}
	return kubeClient, nil
}
