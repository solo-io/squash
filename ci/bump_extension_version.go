package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/solo-io/go-utils/contextutils"
)

// This program does three things:
const (
	// 1. reads and hashes these files
	outputWin32  = "_output/squashctl-windows.exe"
	outputLinux  = "_output/squashctl-linux"
	outputDarwin = "_output/squashctl-darwin"

	// 2. modifies the version field of this file
	jsPackage = "editor/vscode/package.json"

	// 3. produces this file
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

func main() {

	ctx := context.TODO()
	if len(os.Args) != 2 {
		contextutils.LoggerFrom(ctx).Fatal("Must pass a single argument ( version )")
	}
	version := os.Args[1]

	mustPrepareJsPackage(ctx, version)
	mustPrepareConfigFile(ctx, version)

}

func mustPrepareJsPackage(ctx context.Context, version string) {
	lo := contextutils.LoggerFrom(ctx)
	file, err := os.Open(jsPackage)
	if err != nil {
		lo.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	out := ""
	for scanner.Scan() {
		if jsPackageVersionLineMatch.MatchString(scanner.Text()) {
			out += fmt.Sprintf(`  "version": "%v",%v`, version, "\n")
		} else {
			out += scanner.Text() + "\n"
		}
	}
	file.Close()
	writeOut, err := os.Create(jsPackage)
	if err != nil {
		lo.Fatal(err)
	}
	defer writeOut.Close()
	// must be explicit with the format string to prevent accidental substitutions
	if _, err := fmt.Fprintf(writeOut, "%v", out); err != nil {
		lo.Fatal(err)
	}
}

func mustPrepareConfigFile(ctx context.Context, version string) {
	bins := Binaries{
		Linux:  MustGetSha(ctx, outputLinux),
		Darwin: MustGetSha(ctx, outputDarwin),
		Win32:  MustGetSha(ctx, outputWin32),
	}

	squashSpec := SquashSpec{
		Version:  version,
		BaseName: extensionBaseName,
		Binaries: bins,
	}
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

func MustGetSha(ctx context.Context, path string) string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	sha, err := GenerateSoloSha256(file)
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	return sha
}

// TODO(mitchdraft) get this from go-utils
// GenerateSoloSha256 produces a sha256 hash of the given file in a standard way
// for use in Solo.io's build artifact management.
func GenerateSoloSha256(file *os.File) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%v %v\n", hex.EncodeToString(h.Sum(nil)), filepath.Base(file.Name())), nil
}
