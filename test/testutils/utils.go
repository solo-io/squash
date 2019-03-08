package testutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/solo-io/squash/pkg/squashctl"
)

const (
	plankTestVersion = "0.4.7"
	plankTestRepo    = "quay.io/solo-io"
)

func DeclareTestConditions() {
	fmt.Printf(`Squash tests are running under the following conditions:
plank repo: %v
plank tag: %v

If Plank has changed, you should update these values.
`, plankTestRepo, plankTestVersion)
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

func MachineDebugArgs(debugger, ns, podName string) string {
	return fmt.Sprintf(`--debugger %v --machine --namespace %v --pod %v --container-version %v --container-repo %v`, debugger, ns, podName, plankTestVersion, plankTestRepo)
}
