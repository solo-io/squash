package extconfig

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/solo-io/go-utils/contextutils"
)

func MustPrepareReleaseJsPackage(ctx context.Context, version string) {
	lo := contextutils.LoggerFrom(ctx)
	file, err := os.Open(jsPackage)
	if err != nil {
		lo.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	out := ""
	updatedVersion := false
	for scanner.Scan() {
		if jsPackageVersionLineMatch.MatchString(scanner.Text()) {
			out += fmt.Sprintf(`  "version": "%v",%v`, version, "\n")
			updatedVersion = true
		} else {
			out += scanner.Text() + "\n"
		}
	}
	if !updatedVersion {
		lo.Fatal(fmt.Errorf("Did not find version substitution string in package.json, version %v has not been applied", version))
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

func MustPrepareReleaseConfigFile(ctx context.Context, version string) {
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
	mustGenerateConfigFile(ctx, squashSpec)
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
