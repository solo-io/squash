package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/solo-io/squash/pkg/install"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetCmdArgsByPid gets comand line arguments of the running process by PID
func GetCmdArgsByPid(pid int) ([]string, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/cmdline", pid))
	defer f.Close()

	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	l, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}

	s := strings.Replace(string(l), "\x00", " ", -1)
	ss := strings.Split(s, " ")

	return ss, nil
}

func ListSquashDeployments(kc *kubernetes.Clientset, nsList []string) ([]appsv1.Deployment, error) {
	matches := []appsv1.Deployment{}
	for _, ns := range nsList {
		deps, err := kc.Apps().Deployments(ns).List(v1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", install.AgentName)})
		if err != nil {
			return []appsv1.Deployment{}, err
		}
		if len(deps.Items) > 0 {
			matches = append(matches, deps.Items...)
		}
	}
	return matches, nil
}

func DeleteSquashDeployments(kc *kubernetes.Clientset, deps []appsv1.Deployment) (int, error) {
	count := 0
	for _, dep := range deps {
		if err := DeleteSquashDeployment(kc, dep.ObjectMeta.Namespace, dep.ObjectMeta.Name); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func DeleteSquashDeployment(kc *kubernetes.Clientset, namespace, name string) error {
	var gracePeriod int64
	gracePeriod = 0
	err := kc.Apps().Deployments(namespace).Delete(name, &v1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if err != nil {
		return err
	}
	return nil
}

// FindAnyFreePort returns a random port that is not in use.
// It does so by claiming a random open port, then closing it.
func FindAnyFreePort(port *int) error {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return err
	}

	tmpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	*port = tmpListener.Addr().(*net.TCPAddr).Port
	return tmpListener.Close()
}
