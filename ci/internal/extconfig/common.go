package extconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/solo-io/go-utils/contextutils"
)

type SquashSpec struct {
	Version  string   `json:"version"`
	BaseName string   `json:"baseName"`
	Binaries Binaries `json:"binaries"`
}

type Binaries struct {
	Win32  string `json:"win32"`
	Linux  string `json:"linux"`
	Darwin string `json:"darwin"`
}

const (
	osShaIdWin32  = "windows.exe"
	osShaIdLinux  = "linux"
	osShaIdDarwin = "darwin"

	outputWin32  = "_output/squashctl-windows.exe"
	outputLinux  = "_output/squashctl-linux"
	outputDarwin = "_output/squashctl-darwin"

	jsPackage = "editor/vscode/package.json"

	squashConfig = "editor/vscode/src/squash.json"
)

// (End summary of functionality)

const (
	extensionBaseName = "squashctl"
)

var (
	jsPackageVersionLineMatch *regexp.Regexp
)

func init() {
	jsPackageVersionLineMatch = regexp.MustCompile("THIS_VALUE_WILL_BE_REPLACED_BY_BUILD_SCRIPT")
}

func mustGenerateConfigFile(ctx context.Context, squashSpec SquashSpec) {
	json, err := json.MarshalIndent(squashSpec, "", "  ")
	lo := contextutils.LoggerFrom(ctx)
	if err != nil {
		lo.Fatal(err)
	}
	configOut, err := os.Create(squashConfig)
	if err != nil {
		lo.Fatal(err)
	}
	defer configOut.Close()
	if _, err := fmt.Fprintf(configOut, string(json)); err != nil {
		lo.Fatal(err)
	}
}
