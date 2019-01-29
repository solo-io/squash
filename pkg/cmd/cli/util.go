package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
)

func (o *Options) getAllDebugAttachments() (v1.DebugAttachmentList, error) {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	kubeResClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	watchNamespaces, err := kubeutils.GetNamespaces(kubeResClient)
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	das := v1.DebugAttachmentList{}
	for _, ns := range watchNamespaces {
		nsDas, err := (*o.daClient).List(ns, clients.ListOpts{Ctx: o.ctx})
		if err != nil {
			return v1.DebugAttachmentList{}, err
		}
		for _, nsDa := range nsDas {
			das = append(das, nsDa)
		}
	}
	return das, nil
}

func (o *Options) getNamedDebugAttachment(name string) (*v1.DebugAttachment, error) {
	das, err := o.getAllDebugAttachments()
	if err != nil {
		return &v1.DebugAttachment{}, err
	}

	namedDas := v1.DebugAttachmentList{}
	for _, nDa := range das {
		if nDa.Metadata.Name == name {
			namedDas = append(namedDas, nDa)
		}
	}
	if len(namedDas) > 1 {
		// TODO(mitchdraft) - make this impossible by explicitly specifying the namespace
		return &v1.DebugAttachment{}, fmt.Errorf("multiple debug attachments with the same name found")
	}
	if len(namedDas) == 0 {
		return &v1.DebugAttachment{}, fmt.Errorf("Debug attachment %v not found", name)
	}
	return namedDas[0], nil
}

const (
	lcAlpha        = "abcdefghijklmnopqrstuvwxyz"
	lcAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringBytes produces a random string of length n using the characters present in the basis string
func RandStringBytes(n int, basis string) string {
	if basis == "" {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = basis[rand.Intn(len(basis))]
	}
	return string(b)
}

// RandDNS1035 generates a random string of length n that meets the DNS-1035 standard used by Kubernetes names
//
// Typical kubernetes error message for invalid names: a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')

// TODO(mitchdraft) - merge this with go-utils
func RandKubeNameBytes(n int) string {
	if n < 1 {
		return ""
	}
	firstChar := RandStringBytes(1, lcAlpha)
	suffix := ""
	if n > 1 {
		suffix = RandStringBytes(n-1, lcAlphaNumeric)
	}
	return strings.Join([]string{firstChar, suffix}, "")
}

func (top *Options) printVerbose(msg string) {
	if top.Config.verbose {
		fmt.Println(msg)
	}
}

func (top *Options) printVerbosef(tmpl string, args ...interface{}) {
	if top.Config.verbose {
		fmt.Printf(tmpl, args)
	}
}

var logFileName = "cmd.log"

func (top *Options) logCmd(cmd *cobra.Command, args []string) {
	if !top.Config.logCmds {
		return
	}

	cmdWithArgs := fmt.Sprintf("%v %v", cmd.CommandPath(), strings.Join(args, " "))
	flagSpec := getFlagSpec(cmd)
	cmdSpec := fmt.Sprintf("%v %v", cmdWithArgs, flagSpec)

	squashDir, err := squashDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.OpenFile(filepath.Join(squashDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	content := fmt.Sprintf("%v, %v\n", time.Now(), cmdSpec)
	if _, err := f.Write([]byte(content)); err != nil {
		fmt.Println(err)
		return
	}
	if err := f.Close(); err != nil {
		fmt.Println(err)
		return
	}
}

func getChangedFlags(cmd *cobra.Command) map[string]pflag.Value {
	setFlags := make(map[string]pflag.Value)
	ff := func(f *pflag.Flag) {
		if f.Changed {
			fmt.Println("flag changed", f.Name)
			setFlags[f.Name] = f.Value
		}
	}
	cmd.Flags().VisitAll(ff)
	return setFlags
}

func getFlagSpec(cmd *cobra.Command) string {
	flagsChanged := getChangedFlags(cmd)
	str := ""
	for k, v := range flagsChanged {
		switch v.Type() {
		case "bool":
			str += fmt.Sprintf("--%v ", k)
		case "string":
			fallthrough
		default:
			str += fmt.Sprintf("--%v \"%v\" ", k, v)
		}
	}
	return str
}
