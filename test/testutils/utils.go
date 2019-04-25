package testutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/solo-io/build/pkg/ingest"

	"github.com/solo-io/squash/pkg/squashctl"
)

type TestConditions struct {
	PlankImageTag  string
	PlankImageRepo string
	Source         string
}

func InitializeTestConditions(tc *TestConditions) error {
	tc.Source = "build tool"
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
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	_ = w.Close()
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
