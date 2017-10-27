package kubernetes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/utils/podwatcher"

	"os"

	"sync"

	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

type Selector map[string]string
type watchedService struct{}

type KubeOperations struct {
	config *rest.Config

	services     map[string]Selector
	servicesLock sync.RWMutex

	watchcedServices     map[string]watchedService
	watchcedServicesLock sync.RWMutex
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
		config:           config,
		services:         make(map[string]Selector),
		watchcedServices: make(map[string]watchedService),
	}
	go func() {
		for {
			watchfunc, err := podwatcher.Watchservices(ctx, config, s.serviceModified, s.serviceDeleted)
			if err != nil {
				log.Warn("Failed to watch for services.")
				time.Sleep(time.Second)
				continue
			}
			watchfunc()
		}
	}()
	return s, nil
}

func (s *KubeOperations) WatchService(ctx context.Context, servicename string) (<-chan platforms.Container, error) {
	ret := make(chan platforms.Container)

	s.watchcedServicesLock.Lock()
	defer s.watchcedServicesLock.Unlock()
	if _, ok := s.watchcedServices[servicename]; !ok {
		s.watchcedServices[servicename] = watchedService{}
	} else {
		log.WithFields(log.Fields{"service": servicename}).Warn("WatchService: Service is already watched")
		return nil, errors.New("service already being watched")
	}

	selector := s.getServiceSelector(servicename)
	if selector == nil {
		log.WithFields(log.Fields{"service": servicename}).Warn("WatchService: Service has no labels - can't select containers")
		return nil, errors.New("no labels for service")
	}
	go func() {
		podwatcher.Watchpods(ctx, s.config, selector, func(pod *v1.Pod) {
			log.WithFields(log.Fields{"service": servicename, "pod": pod}).Warn("WatchService: new pod for service")
			for _, c := range pod.Spec.Containers {
				ret <- platforms.Container{Name: pod.ObjectMeta.Name + ":" + c.Name, Image: c.Image, Node: pod.Spec.NodeName}
			}
		})
		close(ret)
		s.watchcedServicesLock.Lock()
		defer s.watchcedServicesLock.Unlock()
		delete(s.watchcedServices, servicename)

	}()
	return ret, nil

}
func (s *KubeOperations) getServiceSelector(name string) map[string]string {

	s.servicesLock.RLock()
	defer s.servicesLock.RUnlock()
	return s.services[name]
}

func (s *KubeOperations) serviceModified(name string, selector map[string]string) {
	log.WithFields(log.Fields{"name": name, "selector": selector}).Info("serviceModified")

	s.servicesLock.Lock()
	//if _, ok := s.services[name]; ok {
	s.services[name] = selector
	//}
	s.servicesLock.Unlock()

}

func (s *KubeOperations) serviceDeleted(name string) {
	log.WithFields(log.Fields{"name": name}).Info("serviceDeleted")

	s.servicesLock.Lock()
	delete(s.services, name)
	s.servicesLock.Unlock()

}

func (s *KubeOperations) Locate(context context.Context, containername string) (*platforms.Container, error) {

	clientset, err := kubernetes.NewForConfig(s.config)
	if err != nil {
		log.Warn("Locate - can't get client cluster")
		return nil, err
	}

	var options metav1.GetOptions

	parts := strings.SplitN(containername, ":", 2)
	if len(parts) != 2 {
		log.Warn("Locate - bad name format")
		return nil, errors.New("bad name format")
	}
	podname := parts[0]
	container := parts[1]
	log.WithField("podname", podname).Info("Trying to locate")

	pod, err := clientset.CoreV1().Pods(os.Getenv("KUBE_NAMESPACE")).Get(podname, options)
	if err != nil {
		log.Warn("Locate - can't locate pod ", podname, err)
		return nil, err
	}

	node := pod.Spec.NodeName

	log.WithFields(log.Fields{"podname": podname, "node": node}).Info("Located node for pod")

	newcontainer := &platforms.Container{
		Name: containername,
		Node: node,
	}

	for _, c := range pod.Spec.Containers {
		if c.Name == container {
			newcontainer.Image = c.Image
		}
	}

	return newcontainer, nil
}
