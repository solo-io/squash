package testutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/solo-io/build/pkg/constants"

	"github.com/solo-io/build/pkg/ingest"

	"github.com/solo-io/squash/pkg/squashctl"
)

type TestConditions struct {
	PlankImageTag  string
	PlankImageRepo string
	Source         string
}

func InitializeTestConditions(tc *TestConditions, pathToBuildSpec string) (err error) {
	tc.Source = "build tool"
	initialEnvVar := os.Getenv(constants.EnvVarConfigFileName)
	defer func() {
		if fErr := os.Setenv(constants.EnvVarConfigFileName, initialEnvVar); fErr != nil {
			err = errors.Wrapf(err, "unable to reset env: %v", fErr)
		}
	}()
	// clean this up when https://github.com/solo-io/build/issues/4 is ready
	if err := os.Setenv(constants.EnvVarConfigFileName, pathToBuildSpec); err != nil {
		return err
	}
	buildRun, err := ingest.InitializeBuildRun()
	if err != nil {
		return err
	}
	tc.PlankImageTag = buildRun.Config.ComputedBuildVars.ImageTag
	tc.PlankImageRepo = buildRun.Config.ComputedBuildVars.ContainerPrefix
	if tc.PlankImageTag == "" {
		return fmt.Errorf("unable to read image tag from %s", tc.Source)
	}
	if tc.PlankImageRepo == "" {
		return fmt.Errorf("unable to read container repo from %s", tc.Source)
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

// TOOD - replace with the clicore lib
func Squashctl(args string) error {
	app, err := squashctl.App("test")
	if err != nil {
		return err
	}
	app.SetArgs(strings.Split(args, " "))
	return app.Execute()
}
func SquashctlOut(args string) (string, error) {
	return SquashctlOutWithTimeout(args, nil)
}
func SquashctlOutWithTimeout(args string, timeout *int) (string, error) {
	timeLimit := 600 // default to 10 minute timeout
	if timeout != nil {
		timeLimit = *timeout
	}
	stdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	// back to normal state
	restore := func() {
		_ = w.Close()
		os.Stdout = stdOut // restoring the real stdout
	}
	// have to do this in case it exits by timeout in order to preserve the fail handler output
	// (another reason to use the corecli lib, it writes to buffers during tests, not stdout, so you don't need to
	// risk losing your error messages)
	defer restore()

	app, err := squashctl.App("test")
	if err != nil {
		return "", err
	}
	app.SetArgs(strings.Split(args, " "))

	errC := make(chan error)
	go func() {
		errC <- app.Execute()
	}()
	t := time.NewTimer(time.Duration(timeLimit) * time.Second)

	select {
	case exErr := <-errC:
		if exErr != nil {
			return "", exErr
		}
		break
	case <-t.C:
		return "", fmt.Errorf("timeout during squashctl call")
	}
	outC := make(chan string)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	restore()

	out := <-outC
	return strings.TrimSuffix(out, "\n"), nil
}

func Curl(args string) ([]byte, error) {
	curl := exec.Command("curl", strings.Split(args, " ")...)
	return curl.CombinedOutput()
}

func MachineDebugArgs(tc TestConditions, debugger, ns, podName, squashNamespace, configFile string) string {
	return fmt.Sprintf(`--debugger %v --machine --namespace %v --pod %v --container-version %v --container-repo %v --squash-namespace %v --config %v`,
		debugger,
		ns,
		podName,
		tc.PlankImageTag,
		tc.PlankImageRepo,
		squashNamespace,
		configFile)
}
