package utils

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	squashv1 "github.com/solo-io/squash/pkg/api/v1"
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
		deps, err := kc.Apps().Deployments(ns).List(v1.ListOptions{LabelSelector: fmt.Sprintf("app=%v", install.SquashName)})
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
	return checkAddressAndGetPort(addr, port)
}

// IsPortFree checks if the given port is available for use
func ExpectPortToBeFree(port int) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		return err
	}
	return checkAddressAndGetPort(addr, nil)
}

// if passed an optional port, sets the value to the address's port
func checkAddressAndGetPort(addr *net.TCPAddr, port *int) error {
	tmpListener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	if port != nil {
		*port = tmpListener.Addr().(*net.TCPAddr).Port
	}
	return tmpListener.Close()
}

func ListDebugAttachments(ctx context.Context, daClient squashv1.DebugAttachmentClient, nsList []string) ([]string, error) {
	allDas := []string{}
	for _, ns := range nsList {
		das, err := daClient.List(ns, clients.ListOpts{Ctx: ctx})
		if err != nil {
			return []string{}, err
		}
		for _, da := range das {
			allDas = append(allDas, fmt.Sprintf("%v, %v", ns, da.Metadata.Name))
		}
	}
	return allDas, nil
}

func GetAllDebugAttachments(ctx context.Context, daClient squashv1.DebugAttachmentClient, nsList []string) (squashv1.DebugAttachmentList, error) {
	allDas := squashv1.DebugAttachmentList{}
	for _, ns := range nsList {
		das, err := daClient.List(ns, clients.ListOpts{Ctx: ctx})
		if err != nil {
			return squashv1.DebugAttachmentList{}, err
		}
		for _, da := range das {
			allDas = append(allDas, da)
		}
	}
	return allDas, nil
}
