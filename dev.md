

# Dev workflow notes

## setup a watcher to inspect the debug resources
```
cd test/dev/watcher
go run main
```

## initialize some sample apps and the squash client
```
cd test/dev
go run main --init # to load sample apps and squash client
go run main --att # make an attachment
go run main --clean # remove resources

# whenever you make changes to the squash client (after rebuilding)
go run main --init && go run main --clean
```

## run the e2e tests
```
cd test/e2e
export WAIT_ON_FAIL=1 # if you want better failure debugging
ginkgo -r
```

### run e2e on specific namespaces
```
go run hack/monitor/main.go -namespaces stest-1,stest-2,stest-3,stest-4,stest-5,stest-6
SERIALIZE_NAMESPACES=1 ginkgo -r
```


# Extensions
## Visual Studio Code
- install vsce
```bash
npm install -g vsce
```
- run `publish` from extension's root dir
```bash
vsce publish -p $VSCODE_TOKEN
```



# Using a Pseudoterminal
- The vscode extension connects to debuggers through a proxied port, this lets the proxy detect when the connection closes and then "complete" the debugger pod.
- To get this behavior on the terminal interface, we need to use a pseudoterminal. Otherwise, debugger control signals are interpreted as proxy termination requests.
- In basic testing, the pty terminal seemed like a good candidate. There were some quirks, more testing is needed to iron this out.
- For now, just connect to the debug port, not the proxied port.
- In the future, the snippet below may be helpful:
```diff
diff --git a/pkg/config/squash.go b/pkg/config/squash.go
index 5e736cd..7602a60 100644
--- a/pkg/config/squash.go
+++ b/pkg/config/squash.go
@@ -6,11 +6,15 @@ import (
 	"io"
 	"os"
 	"os/exec"
+	"os/signal"
 	"strings"
+	"syscall"
 	"time"
 
+	"github.com/kr/pty"
 	"github.com/pkg/errors"
 	log "github.com/sirupsen/logrus"
+	"golang.org/x/crypto/ssh/terminal"
 
 	gokubeutils "github.com/solo-io/go-utils/kubeutils"
 	sqOpts "github.com/solo-io/squash/pkg/options"
@@ -137,7 +141,7 @@ func (s *Squash) connectUser(createdPod *v1.Pod) error {
 		return nil
 	}
 	// Starting port forward in background.
-	portSpec := fmt.Sprintf("%v:%v", s.LocalPort, sqOpts.DebuggerPort)
+	portSpec := fmt.Sprintf("%v:%v", s.LocalPort, sqOpts.OutPort)
 	cmd1 := exec.Command("kubectl", "port-forward", createdPod.ObjectMeta.Name, portSpec, "-n", s.getDebuggerPodNamespace())
 	cmd1.Stdout = os.Stdout
 	cmd1.Stderr = os.Stderr
@@ -157,16 +161,62 @@ func (s *Squash) connectUser(createdPod *v1.Pod) error {
 
 	// TODO(mitchdraft) dlv only atm - check if dlv before doing this
 	cmd2 := exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", s.LocalPort))
-	cmd2.Stdout = os.Stdout
-	cmd2.Stderr = os.Stderr
-	cmd2.Stdin = os.Stdin
-	err = cmd2.Run()
+	err = ptyWrap(cmd2)
 	if err != nil {
 		log.Warn("failed, printing logs")
 		log.Warn(err)
 		s.showLogs(err, createdPod)
 		return err
 	}
+	// cmd2 := exec.Command("dlv", "connect", fmt.Sprintf("127.0.0.1:%v", s.LocalPort))
+	// cmd2.Stdout = os.Stdout
+	// cmd2.Stderr = os.Stderr
+	// cmd2.Stdin = os.Stdin
+	// err = cmd2.Run()
+	// if err != nil {
+	// 	log.Warn("failed, printing logs")
+	// 	log.Warn(err)
+	// 	s.showLogs(err, createdPod)
+	// 	return err
+	// }
+	return nil
+}
+
+func ptyWrap(c *exec.Cmd) error {
+	// Create arbitrary command.
+	// c := exec.Command("bash")
+
+	// Start the command with a pty.
+	ptmx, err := pty.Start(c)
+	if err != nil {
+		return err
+	}
+	// Make sure to close the pty at the end.
+	defer func() { _ = ptmx.Close() }() // Best effort.
+
+	// Handle pty size.
+	ch := make(chan os.Signal, 1)
+	signal.Notify(ch, syscall.SIGWINCH)
+	go func() {
+		for range ch {
+			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
+				log.Printf("error resizing pty: %s", err)
+			}
+		}
+	}()
+	ch <- syscall.SIGWINCH // Initial resize.
+
+	// Set stdin in raw mode.
+	oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
+	if err != nil {
+		panic(err)
+	}
+	defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.
+
+	// Copy stdin to the pty and the pty to stdout.
+	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
+	_, _ = io.Copy(os.Stdout, ptmx)
+
 	return nil
 }
 
```
