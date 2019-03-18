package squashctl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/cliutils"
	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/actions"
	"github.com/solo-io/squash/pkg/config"
	"github.com/solo-io/squash/pkg/options"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
	squashkubeutils "github.com/solo-io/squash/pkg/utils/kubeutils"
	"github.com/solo-io/squash/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/AlecAivazis/survey.v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

/*
Notes on CLI design

An options struct is populated by a combination of:
- user input args
- user input flags
- env variables
- config file
- defaults

A specific command is specified by a chain of strings

The options struct is interpreted according to the command
Ideally, the options struct's format should follow the command tree's format

All commands should have an interactive mode.
Interactive mode and option validation can be implemented with this pattern:
```
if err := top.ensureParticularCmdOption(po *particularOption); err != nil {
    return err
}
```
- Methods should be built off of the root of the options tree (the "top" var in the example above). This allows sub commands to share common values.
- Sub commands should only modify their portion of the options tree. (This makes it easier to move sub commands around if we want a different organization later).

*/

const descriptionUsage = `Squash requires no arguments. Just run it!
It creates a privileged debug pod, starts a debugger, and then attaches to it.
If you are debugging in a shared cluster, consider using Squash (in cluster process).
(squashctl squash --help for more info)
Find more information at https://solo.io
`

func App(version string) (*cobra.Command, error) {
	opts := NewOptions()
	app := &cobra.Command{
		Use:     "squashctl",
		Short:   "debug microservices with squash",
		Long:    descriptionUsage,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			opts.readConfigValues(&opts.Config)
			opts.logCmd(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// when no sub commands are specified, run w/wo RBAC according to settings
			return opts.runBaseCommand()
		},
		SuggestionsMinimumDistance: 1,
	}

	if err := initializeOptions(opts); err != nil {
		return &cobra.Command{}, err
	}

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		opts.DeployCmd(opts),
		opts.SquashCmd(opts),
		opts.UtilsCmd(opts),
		completionCmd(),
	)

	app.PersistentFlags().BoolVar(&opts.Json, "json", false, "output json format")
	applySquashFlags(&opts.Squash, app.PersistentFlags())

	return app, nil
}

func applySquashFlags(cfg *config.Squash, f *pflag.FlagSet) {
	f.BoolVar(&cfg.NoClean, "no-clean", false, "don't clean temporary pod when existing")
	f.BoolVar(&cfg.ChooseDebugger, "no-guess-debugger", false, "don't auto detect debugger to use")
	f.BoolVar(&cfg.ChoosePod, "no-guess-pod", false, "don't auto detect pod to use")
	f.IntVar(&cfg.TimeoutSeconds, "timeout", 300, "timeout in seconds to wait for debug pod to be ready")
	f.StringVar(&cfg.DebugContainerVersion, "container-version", version.ImageVersion, "debug container version to use")
	f.StringVar(&cfg.DebugContainerRepo, "container-repo", version.ImageRepo, "debug container repo to use")

	f.IntVar(&cfg.LocalPort, "localport", 0, "local port to use to connect to debugger (defaults to random free port)")

	f.BoolVar(&cfg.Machine, "machine", false, "machine mode input and output")
	f.StringVar(&cfg.Debugger, "debugger", "", "Debugger to use")
	f.StringVar(&cfg.Namespace, "namespace", "", "Namespace to debug")
	f.StringVar(&cfg.Pod, "pod", "", "Pod to debug")
	f.StringVar(&cfg.Container, "container", "", "Container to debug")
	f.StringVar(&cfg.CRISock, "crisock", "/var/run/dockershim.sock", "The path to the CRI socket")
	f.StringVar(&cfg.SquashNamespace, "squash-namespace", sqOpts.SquashNamespace, fmt.Sprintf("the namespace where squash resources will be deployed (default: %v)", options.SquashNamespace))
}

func initializeOptions(o *Options) error {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return err
	}
	o.ctx = ctx
	o.daClient = daClient

	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return err
	}
	o.KubeClient = kubeClient

	o.DeployOptions = defaultDeployOptions()
	return nil
}

func (o *Options) runBaseCommand() error {
	o.printVerbose("Attaching debugger")

	if err := o.ensureMinimumSquashConfig(); err != nil {
		return err
	}

	if err := o.cleanupPreRun(); err != nil {
		return err
	}

	if o.Config.secureMode {
		o.printVerbose("squashctl will create a CRD with your debug intent in your target pod's namespace. Squash will create a debugger pod in your target pod's.")
		if err := o.runBaseCommandWithRbac(); err != nil {
			o.cleanupPostRun()
			return err
		}
	} else {
		o.printVerbose("Squash will create a debugger pod in your target pod's namespace.")
		if err := o.createPlankPermissions(); err != nil {
			return err
		}
		// TODO - replace sleeps with watches on CRD
		time.Sleep(200 * time.Millisecond)
		if err := o.writeDebugAttachment(); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
		_, err := config.StartDebugContainer(o.Squash, o.DebugTarget)
		o.cleanupPostRun()
		return err
	}
	o.cleanupPostRun()

	return nil
}

func (top *Options) runBaseCommandWithRbac() error {
	if err := top.ensureSquashIsInCluster(); err != nil {
		return err
	}

	if err := top.createPlankPermissions(); err != nil {
		return err
	}

	if err := top.writeDebugAttachment(); err != nil {
		return err
	}

	// wait until pod is created, print its name so the extension can connect

	// TODO(mitchdraft) - add this to the configuration file
	// 1 second was not long enough, status still pending, could not port-forward
	// 3 seconds might be overkill
	// TODO(mitchdraft) - replace with watch on cmd stream
	time.Sleep(3 * time.Second)

	return top.Squash.ReportOrConnectToCreatedDebuggerPod()
}

func (o *Options) writeDebugAttachment() error {
	so := o.Squash
	dbge := o.DebugTarget

	uc, err := actions.NewUserController()
	if err != nil {
		return err
	}
	daName := cliutils.RandKubeNameBytes(10)
	// this works in the form: `squash  --namespace mk6 --pod example-service1-74bbc5dcd-rvrtq`
	_, err = uc.Attach(
		daName,
		so.Namespace,
		dbge.Container.Image,
		dbge.Pod.Name,
		so.Container,
		"",
		so.Debugger)

	return nil
}

func (o *Options) ensureMinimumSquashConfig() error {

	if err := o.chooseDebugger(); err != nil {
		return err
	}
	if err := o.GetMissing(); err != nil {
		return err
	}
	if err := o.ensureLocalPort(&o.Squash.LocalPort); err != nil {
		return err
	}

	if !o.Squash.Machine {
		confirmed := false
		prompt := &survey.Confirm{
			Message: "Going to attach " + o.Squash.Debugger + " to pod " + o.DebugTarget.Pod.ObjectMeta.Name + ". continue?",
			Default: true,
		}
		survey.AskOne(prompt, &confirmed, nil)
		if !confirmed {
			return errors.New("user aborted")
		}
	}
	return nil
}

func (o *Options) chooseDebugger() error {
	if o.Squash.Debugger != "" {
		return nil
	}

	debugger := o.detectLang()

	if debugger == "" {
		question := &survey.Select{
			Message: "Select a debugger",
			Options: sqOpts.AvailableDebuggers,
		}
		var choice string
		if err := survey.AskOne(question, &choice, survey.Required); err != nil {
			return err
		}
		debugger = choice
	}
	o.Squash.Debugger = debugger
	return nil
}

func (o *Options) detectLang() string {
	if o.Squash.ChooseDebugger {
		// manual mode
		return ""
	}
	// TODO: find some decent huristics to make this work
	return ""
}

func (o *Options) GetMissing() error {

	//	clientset.CoreV1().Namespace().
	// see if namespace exist, and if not prompt for one.
	if o.Squash.Namespace == "" {
		if err := o.chooseAllowedNamespace(&(o.Squash.Namespace), "Select a namespace to debug"); err != nil {
			return errors.Wrap(err, "choosing namespace")
		}
	}

	if o.Squash.Pod == "" {
		if err := o.choosePod(); err != nil {
			return errors.Wrap(err, "choosing pod")
		}
	} else {
		if err := o.Squash.GetDebugTargetPodFromSpec(&o.DebugTarget); err != nil {
			return err
		}
	}

	if o.Squash.Container == "" {
		if err := chooseContainer(o); err != nil {
			return errors.Wrap(err, "choosing container")
		}
	} else {
		if err := o.Squash.GetDebugTargetContainerFromSpec(&o.DebugTarget); err != nil {
			return err
		}
	}
	return nil
}

func chooseContainer(o *Options) error {
	pod := o.DebugTarget.Pod
	if len(pod.Spec.Containers) == 0 {
		return errors.New("no container to choose from")

	}
	if len(pod.Spec.Containers) == 1 {
		o.DebugTarget.Container = &pod.Spec.Containers[0]
		o.Squash.Container = pod.Spec.Containers[0].Name
		return nil
	}

	containerNames := make([]string, 0, len(pod.Spec.Containers))
	for _, container := range pod.Spec.Containers {
		contname := container.Name
		containerNames = append(containerNames, contname)
	}

	question := &survey.Select{
		Message: "Select a container",
		Options: containerNames,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		return err
	}

	for _, container := range pod.Spec.Containers {
		if choice == container.Name {
			o.DebugTarget.Container = &container
			o.Squash.Container = container.Name
			return nil
		}
	}

	return errors.New("selected container not found")
}

func (o *Options) chooseAllowedNamespace(target *string, question string) error {

	namespaces, err := o.KubeClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "reading namespaces")
	}
	namespaceNames := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		nsname := ns.ObjectMeta.Name
		if nsname == "squash" {
			continue
		}
		if strings.HasPrefix(nsname, "kube-") {
			continue
		}
		namespaceNames = append(namespaceNames, nsname)
	}
	if len(namespaceNames) == 0 {
		return errors.New("no namespaces available!")
	}

	if len(namespaceNames) == 1 {
		*target = namespaceNames[0]
		return nil
	}

	prompt := &survey.Select{
		Message: question,
		Options: namespaceNames,
	}
	if err := survey.AskOne(prompt, target, survey.Required); err != nil {
		return err
	}
	return nil
}

func (o *Options) choosePod() error {

	pods, err := o.KubeClient.CoreV1().Pods(o.Squash.Namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "reading namesapces")
	}
	podName := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if o.Squash.ChoosePod || o.Squash.Container == "" {
			podName = append(podName, pod.ObjectMeta.Name)
		} else {
			for _, podContainer := range pod.Spec.Containers {
				if strings.HasPrefix(podContainer.Image, o.Squash.Container) {
					podName = append(podName, pod.ObjectMeta.Name)
					break
				}
			}
		}
	}

	var choice string
	if len(podName) == 1 {
		choice = podName[0]
	} else {
		question := &survey.Select{
			Message: "Select a pod",
			Options: podName,
		}
		if err := survey.AskOne(question, &choice, survey.Required); err != nil {
			return err
		}
	}
	for _, pod := range pods.Items {
		if choice == pod.ObjectMeta.Name {
			o.DebugTarget.Pod = &pod
			o.Squash.Pod = pod.ObjectMeta.Name
			return nil
		}
	}

	return errors.New("pod not found")
}

func (o *Options) ensureLocalPort(port *int) error {
	if port == nil {
		return fmt.Errorf("Port must not be nil")
	}
	if *port == 0 {
		// In this case, user wants to use a random open port.
		// We need to know the port so we can configure port-forwarding
		// so rather than letting the os choose an unknown port we
		// find a port that we know to be free.
		if err := utils.FindAnyFreePort(port); err != nil {
			return err
		}
	} else {
		if err := utils.ExpectPortToBeFree(*port); err != nil {
			return fmt.Errorf("Port %v already in use. Please choose a different port or remove the --localport flag for a free port to be chosen automatically.", *port)
		}
	}
	return nil
}

func (o *Options) ensureSquashIsInCluster() error {
	nsList, err := squashkubeutils.GetNamespaces(o.KubeClient)
	if err != nil {
		return err
	}
	squashDeployments, err := utils.ListSquashDeployments(o.KubeClient, nsList)
	if err != nil {
		return err
	}

	if len(squashDeployments) == 0 {
		return fmt.Errorf("Squash must be deployed to the cluster to use secure mode. Either disable secure mode in your squash config file or deploy Squash to your cluster. You can deploy with 'squashctl squash deploy'.")
	}

	return nil
}

func (o *Options) cleanupPreRun() error {
	// look for an existing Debug Attachment CRD and clean up its old resources
	it := o.Squash.GetIntent()
	priorDas, err := it.GetDebugAttachments(o.daClient)
	if err != nil {
		return errors.Wrap(err, "cleanup pre run list das")
	}
	for _, priorDa := range priorDas {
		// delete the prior plank pod
		if err := o.KubeClient.
			CoreV1().
			Pods(o.Squash.SquashNamespace).
			Delete(priorDa.PlankName, &meta_v1.DeleteOptions{}); err != nil {
			// do not exit on error, it does not matter if the pod was already deleted
			// TODO(mitchdraft) - first check if the pod exists before deleting
			if !o.Squash.Machine {
				fmt.Println(err)
			}
		}
		if err := o.daClient.Delete(
			priorDa.Metadata.Namespace,
			priorDa.Metadata.Name,
			clients.DeleteOpts{}); err != nil {
			if !o.Squash.Machine {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func (o *Options) cleanupPostRun() error {
	// remove pod only if we are not in machine mode
	if !o.Squash.Machine {
		it := o.Squash.GetIntent()
		priorDa, err := it.GetDebugAttachment(o.daClient)
		if err != nil {
			return errors.Wrap(err, "cleanup pre run list das")
		}
		if err := o.daClient.Delete(
			priorDa.Metadata.Namespace,
			priorDa.Metadata.Name,
			clients.DeleteOpts{}); err != nil {
			fmt.Println(err)
		}

		if err := o.Squash.DeletePlankPod(); err != nil {
			return err
		}
	}

	return nil
}
