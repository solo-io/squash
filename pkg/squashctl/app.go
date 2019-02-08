package squashctl

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/actions"
	"github.com/solo-io/squash/pkg/kscmd"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/spf13/cobra"
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
If you are debugging in a shared cluster, consider using squash the squash agent.
(squash agent --help for more info)
Find more information at https://solo.io
`

func App(version string) (*cobra.Command, error) {
	opts := &Options{}
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
	}

	if err := initializeOptions(opts); err != nil {
		return &cobra.Command{}, err
	}

	app.SuggestionsMinimumDistance = 1
	app.AddCommand(
		DebugContainerCmd(opts),
		DebugRequestCmd(opts),
		ListCmd(opts),
		WaitAttCmd(opts),
		opts.DeployCmd(opts),
		opts.AgentCmd(opts),
	)

	app.PersistentFlags().BoolVar(&opts.Json, "json", false, "output json format")
	applyLiteFlags(&opts.LiteOptions, app.PersistentFlags())

	return app, nil
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

	// WIP - pausing for UI - this will gather the interactive stuff
	if err := o.ensureMinimumSquashConfig(); err != nil {
		return err
	}

	if o.Config.secureMode {
		o.printVerbose("Squash will create a CRD with your debug intent in your target pod's namespace. The squash agent will create a debugger pod in your target pod's.")
		if err := o.runBaseCommandWithRbac(); err != nil {
			return err
		}
	} else {
		o.printVerbose("Squash will create a debugger pod in your target pod's namespace.")
		_, err := kscmd.StartDebugContainer(o.LiteOptions, o.Debugee)
		return err
	}

	return nil
}

func (top *Options) runBaseCommandWithRbac() error {
	uc, err := actions.NewUserController()
	if err != nil {
		return err
	}

	// TODO(mitchdraft) - use kubernetes' generate name instead of making a dummy name
	daName := fmt.Sprintf("da-%v", rand.Int31n(100000))

	lo := top.LiteOptions

	podSpec, err := top.KubeClient.Core().Pods(lo.Namespace).Get(lo.Pod, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	// TODO(mitchdraft) - choose among images (rather than taking the first)
	image := podSpec.Spec.Containers[0].Image

	// this works in the form: `squash  --namespace mk6 --pod example-service1-74bbc5dcd-rvrtq`
	// TODO(mitchdraft) - get these values interactively
	_, err = uc.Attach(
		daName,
		lo.Namespace,
		image,
		lo.Pod,
		lo.Container,
		"",
		lo.Debugger)
	return err
}

func (o *Options) ensureMinimumSquashConfig() error {

	if err := o.chooseDebugger(); err != nil {
		return err
	}
	if err := o.GetMissing(); err != nil {
		return err
	}

	if !o.Squash.Machine {
		confirmed := false
		prompt := &survey.Confirm{
			Message: "Going to attach " + o.Squash.Debugger + " to pod " + o.Debugee.Pod.ObjectMeta.Name + ". continue?",
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
	if len(o.Squash.Debugger) != 0 {
		return nil
	}

	availableDebuggers := []string{"dlv", "gdb"}
	debugger := o.detectLang()

	if debugger == "" {
		question := &survey.Select{
			Message: "Select a debugger",
			Options: availableDebuggers,
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
	return "dlv"
}

func (o *Options) GetMissing() error {

	//	clientset.CoreV1().Namespace().
	// see if namespace exist, and if not prompt for one.
	if o.Squash.Namespace == "" {
		if err := o.chooseNamespace(); err != nil {
			return errors.Wrap(err, "choosing namespace")
		}
	}

	if o.Squash.Pod == "" {
		if err := o.choosePod(); err != nil {
			return errors.Wrap(err, "choosing pod")
		}
	} else {
		var err error
		o.Debugee.Pod, err = o.KubeClient.CoreV1().Pods(o.Squash.Namespace).Get(o.Squash.Pod, meta_v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "fetching pod")
		}
	}

	if o.Squash.Container == "" {
		if err := chooseContainer(o); err != nil {
			return errors.Wrap(err, "choosing container")
		}
	} else {
		for _, podContainer := range o.Debugee.Pod.Spec.Containers {
			log.Debug(podContainer.Image)
			if strings.HasPrefix(podContainer.Image, o.Squash.Container) {
				o.Debugee.Container = &podContainer
				break
			}
		}
		if o.Debugee.Container == nil {
			// time.Sleep(555 * time.Second)
			return errors.New(fmt.Sprintf("no such container image: %v", o.Squash.Container))
		}
	}
	return nil
}

func chooseContainer(o *Options) error {
	pod := o.Debugee.Pod
	if len(pod.Spec.Containers) == 0 {
		return errors.New("no container to choose from")

	}
	if len(pod.Spec.Containers) == 1 {
		o.Debugee.Container = &pod.Spec.Containers[0]
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
			o.Debugee.Container = &container
			return nil
		}
	}

	return errors.New("selected container not found")
}

func (o *Options) chooseNamespace() error {

	namespaces, err := o.KubeClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "reading namesapces")
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
		o.Squash.Namespace = namespaceNames[0]
		return nil
	}

	question := &survey.Select{
		Message: "Select a namespace",
		Options: namespaceNames,
	}
	if err := survey.AskOne(question, &o.Squash.Namespace, survey.Required); err != nil {
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
			o.Debugee.Pod = &pod
			return nil
		}
	}

	return errors.New("pod not found")
}
