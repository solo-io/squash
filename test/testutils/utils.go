package testutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/solo-io/squash/pkg/squashctl"
)

const (
	plankImageTagEnvVar  = "PLANK_IMAGE_TAG"
	plankImageRepoEnvVar = "PLANK_IMAGE_REPO"
	buildOutputFilepath  = "../../_output/buildtimevalues.yaml"
)

type TestConditions struct {
	PlankImageTag  string `yaml:"plank_image_tag"`
	PlankImageRepo string `yaml:"plank_image_repo"`
	Source         string `yaml:"source"`
}

func getBuildValue(filepath string, tc *TestConditions) error {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(content, tc); err != nil {
		return err
	}
	return nil
}

func InitializeTestConditionsFromBuildTimeFile(tc *TestConditions) error {
	tc.Source = fmt.Sprintf("build time values in %s", buildOutputFilepath)
	if err := getBuildValue(buildOutputFilepath, tc); err != nil {
		return err
	}
	if tc.PlankImageTag == "" {
		return fmt.Errorf("must set plank_image_tag in %s", buildOutputFilepath)
	}
	if tc.PlankImageRepo == "" {
		return fmt.Errorf("must set plank_image_repo in %s", buildOutputFilepath)
	}
	return nil
}

func InitializeTestConditionsFromEnv(tc *TestConditions) error {
	tc.Source = "environment variables"
	tc.PlankImageTag = os.Getenv(plankImageTagEnvVar)
	tc.PlankImageRepo = os.Getenv(plankImageRepoEnvVar)
	if tc.PlankImageTag == "" {
		return fmt.Errorf("must set %s env var", plankImageTagEnvVar)
	}
	if tc.PlankImageRepo == "" {
		return fmt.Errorf("must set %s env var", plankImageRepoEnvVar)
	}
	return nil
}

func SummarizeTestConditions(tc TestConditions) string {
	return fmt.Sprintf(`Squash tests are running under the following conditions:
plank repo: %s
plank tag: %s
values set from %s
`, tc.PlankImageRepo, tc.PlankImageTag, tc.Source)
}

func Squashctl(args string) error {
	app, err := squashctl.App("test")
	if err != nil {
		return err
	}
	app.SetArgs(strings.Split(args, " "))
	return app.Execute()
}
func SquashctlOut(args string) (string, error) {
	stdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	app, err := squashctl.App("test")
	if err != nil {
		return "", err
	}
	app.SetArgs(strings.Split(args, " "))
	err = app.Execute()

	outC := make(chan string)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = stdOut // restoring the real stdout
	out := <-outC

	return strings.TrimSuffix(out, "\n"), nil
}

func Curl(args string) ([]byte, error) {
	curl := exec.Command("curl", strings.Split(args, " ")...)
	return curl.CombinedOutput()
}

func MachineDebugArgs(tc TestConditions, debugger, ns, podName, squashNamespace string) string {
	return fmt.Sprintf(`--debugger %v --machine --namespace %v --pod %v --container-version %v --container-repo %v --squash-namespace %v`,
		debugger,
		ns,
		podName,
		tc.PlankImageTag,
		tc.PlankImageRepo,
		squashNamespace)
}
