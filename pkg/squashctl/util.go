package squashctl

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/AlecAivazis/survey.v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	daClient, err := o.getDAClient()
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	for _, ns := range watchNamespaces {
		nsDas, err := daClient.List(ns, clients.ListOpts{Ctx: o.ctx})
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
		return nil, err
	}

	namedDas := v1.DebugAttachmentList{}
	for _, nDa := range das {
		if nDa.Metadata.Name == name {
			namedDas = append(namedDas, nDa)
		}
	}
	if len(namedDas) > 1 {
		// TODO(mitchdraft) - make this impossible by explicitly specifying the namespace
		return nil, fmt.Errorf("multiple debug attachments with the same name found")
	}
	if len(namedDas) == 0 {
		return nil, fmt.Errorf("Debug attachment %v not found", name)
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

func (o *Options) printVerbose(msg string) {
	if o.Config.verbose && !o.Squash.Machine {
		fmt.Println(msg)
	}
}

func (o *Options) printVerbosef(tmpl string, args ...interface{}) {
	if o.Config.verbose {
		fmt.Printf(tmpl, args)
	}
}

var logFileName = "cmd.log"

func (o *Options) logCmd(cmd *cobra.Command, args []string) {
	if !o.Config.logCmds {
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

func (o *Options) chooseString(message string, choice *string, options []string) error {
	question := &survey.Select{
		Message: message,
		Options: options,
	}
	if err := survey.AskOne(question, choice, survey.Required); err != nil {
		return err
	}
	return nil
}

// TODO - call this once from squashctl once during config instead of each time a pod is created
// for now - just print errors since these resources may have already been created
// we need them to exist in each namespace
func (o *Options) createPlankPermissions() error {

	cs, err := getClientSet()
	if err != nil {
		return err
	}

	// need to create the permissions in the namespace of the target process
	namespace := o.Squash.SquashNamespace

	if _, err := cs.CoreV1().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: sqOpts.PlankServiceAccountName,
		},
	}

	o.info(fmt.Sprintf("Creating service account %v in namespace %v\n", sqOpts.PlankServiceAccountName, namespace))
	if _, err := cs.CoreV1().ServiceAccounts(namespace).Create(&sa); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: sqOpts.PlankClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
				Resources: []string{"pods"},
				APIGroups: []string{""},
			},
			{
				Verbs:     []string{"list"},
				Resources: []string{"namespaces"},
				APIGroups: []string{""},
			},
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
				Resources: []string{"debugattachments"},
				APIGroups: []string{"squash.solo.io"},
			},
			{
				// TODO remove the register permission when solo-kit is updated
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete", "register"},
				Resources: []string{"customresourcedefinitions"},
				APIGroups: []string{"apiextensions.k8s.io"},
			},
		},
	}
	o.info(fmt.Sprintf("Creating cluster role %v \n", sqOpts.PlankClusterRoleName))
	if _, err := cs.Rbac().ClusterRoles().Create(cr); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sqOpts.PlankClusterRoleBindingName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sqOpts.PlankServiceAccountName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Name: sqOpts.PlankClusterRoleName,
			Kind: "ClusterRole",
		},
	}

	o.info(fmt.Sprintf("Creating cluster role binding %v \n", sqOpts.PlankClusterRoleBindingName))
	if _, err := cs.Rbac().ClusterRoleBindings().Create(crb); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}
	o.info(fmt.Sprintf("All squashctl permission resources created.\n"))
	return nil
}

func getClientSet() (kubernetes.Interface, error) {
	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restCfg)
}

// only print info if squashctl is being used by a human
// machine mode currently expects an exact output
func (o *Options) info(msg string) {
	if !o.Squash.Machine {
		fmt.Println(msg)
	}
}
