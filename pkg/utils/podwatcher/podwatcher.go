package podwatcher

import (
	"context"

	"os"

	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

// watch service list
// for each service watch the selectors
// watch pods with the selectors

func Watchservices(ctx context.Context, config *rest.Config, modify func(n string, selector map[string]string),
	del func(n string)) (func(), error) {

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Warn("Watchservices - can't get client cluster")
		return nil, err
	}

	options := metav1.ListOptions{
		Watch: true,
	}

	w, err := clientset.CoreV1().Services(os.Getenv("KUBE_NAMESPACE")).Watch(options)
	if err != nil {
		log.Warn("Watchservices - can't watch services")
		return nil, err
	}

	return func() {

		cancel := stopwatching(ctx, w)
		defer cancel()
		log.Info("Watchservices - starting watch")
		defer log.Info("Watchservices -  watch ended")

		for e := range w.ResultChan() {
			switch e.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				s := e.Object.(*v1.Service)
				// TODO: instead of storing config in memory, store it in the cluster as labels.
				// if s.Labels["solo.io"] == "debug" {
				modify(s.Name, s.Spec.Selector)
				// } else {
				// 	del(s.Name)
				// }
			case watch.Deleted:
				s := e.Object.(*v1.Service)
				del(s.Name)
			}
		}
	}, nil

}

func stopwatching(ctx context.Context, w watch.Interface) func() {
	donewatching := make(chan struct{}, 1)

	go func() {
		select {
		case <-ctx.Done():
			w.Stop()
		case <-donewatching:
		}
	}()
	return func() { donewatching <- struct{}{} }
}

func Watchpods(ctx context.Context, config *rest.Config, selector map[string]string, added func(*v1.Pod)) error {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Warn("AttachContainer - can't get client cluster")
		return err
	}
	// TODO: try to make the selector more future proof
	var options metav1.ListOptions
	for k, v := range selector {
		if options.LabelSelector != "" {
			options.LabelSelector += ","
		}
		options.LabelSelector += k + "=" + v

	}
	w, err := clientset.CoreV1().Pods(os.Getenv("KUBE_NAMESPACE")).Watch(options)
	if err != nil {
		log.Warn("AttachContainer - can't watch services")
		return err
	}

	cancel := stopwatching(ctx, w)
	defer cancel()

	for e := range w.ResultChan() {
		switch e.Type {
		case watch.Added:
			fallthrough
		case watch.Modified:
			p := e.Object.(*v1.Pod)
			if p.Spec.NodeName == "" {
				continue
			}
			if p.Status.Phase != v1.PodRunning {
				continue
			}
			added(p)
		}
	}

	return nil
}
