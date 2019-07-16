package kubernetes

import (
	"context"
	"errors"

	"github.com/solo-io/go-utils/contextutils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/solo-io/squash/pkg/platforms"
	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
)

type KubeOperations struct {
	config *rest.Config
}

func NewKubeOperations(ctx context.Context, config *rest.Config) (*KubeOperations, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if config == nil {
		inconfig, err := rest.InClusterConfig()
		if err != nil {
			panic("not running in a kube cluster")
		}
		config = inconfig
	}
	s := &KubeOperations{
		config: config,
	}
	return s, nil
}

func (s *KubeOperations) Locate(context context.Context, attachment interface{}) (interface{}, *platforms.Container, error) {

	kubeAttachment, err := k8models.GenericToKubeAttachment(attachment)
	logger := contextutils.LoggerFrom(context)
	if err != nil {
		logger.Warn("Locate - error converting attachment")
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(s.config)
	if err != nil {
		logger.Warn("Locate - can't get client cluster")
		return nil, nil, err
	}

	var options metav1.GetOptions

	if kubeAttachment.Namespace == "" {
		kubeAttachment.Namespace = "default"
	}

	logger.Infow("Trying to locate", "podname", kubeAttachment.Pod)

	pod, err := clientset.CoreV1().Pods(kubeAttachment.Namespace).Get(kubeAttachment.Pod, options)
	if err != nil {
		logger.Warn("Locate - can't locate pod ", kubeAttachment.Pod, err)
		return nil, nil, err
	}

	node := pod.Spec.NodeName

	logger.Infow("Located node for pod", "podname", kubeAttachment.Pod, "node", node)

	newcontainer := &platforms.Container{
		Name: kubeAttachment.Container,
		Node: node,
	}

	if newcontainer.Name != "" {
		for _, c := range pod.Spec.Containers {
			if c.Name == newcontainer.Name {
				newcontainer.Image = c.Image
			}
		}
	} else {
		// find the relevant container
		if len(pod.Spec.Containers) == 1 {
			// easy case
			c := pod.Spec.Containers[0]
			newcontainer.Name = c.Name
			newcontainer.Image = c.Image
		} else {
			// filter to only containers with ports
			var potentialContainers []v1.Container
			for _, c := range pod.Spec.Containers {
				if len(c.Ports) > 0 {
					potentialContainers = append(potentialContainers, c)
				}
			}

			if len(potentialContainers) == 1 {
				c := potentialContainers[0]
				newcontainer.Name = c.Name
				newcontainer.Image = c.Image
			} else {
				logger.Warnf("Couldn't determine which container we need to debug", "potentialContainers", potentialContainers)
				return nil, nil, errors.New("cant find container to debug")
			}
		}

	}
	kubeAttachment.Container = newcontainer.Name

	return kubeAttachment, newcontainer, nil
}
