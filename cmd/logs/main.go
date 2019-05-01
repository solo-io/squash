package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	skube "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"
	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {
	cs, err := getClientset()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := dumpLogs(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(4 * time.Second)
	cancel()
	fmt.Println("just cancelled")
	time.Sleep(4 * time.Second)
	fmt.Println("now done")
}

func dumpLogs(ctx context.Context) error {
	ps := &podFilter{name: "tmp"}
	arts := []*v1alpha2.Artifact{{
		ImageName: "",
		Workspace: "",
	}}
	cp := skube.NewColorPicker(arts)
	la := skube.NewLogAggregator(os.Stdout, ps, cp)
	fmt.Println("about to start agg")
	if err := la.Start(ctx); err != nil {
		return err
	}
	return nil
}

type podFilter struct {
	name string
}

func (pf podFilter) Select(pod *v1.Pod) bool {
	return true // for now anyway
	//if pod.Name == pf.name {
	//	return true
	//}
	//return false
}

func getClientset() (*kubernetes.Clientset, error) {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restCfg)
}
