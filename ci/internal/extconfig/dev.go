package extconfig

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
)

func MustCreateDevResources(ctx context.Context, version string) {
	bins := Binaries{
		Linux:  MustDownloadSha(ctx, version, osShaIdLinux),
		Darwin: MustDownloadSha(ctx, version, osShaIdDarwin),
		Win32:  MustDownloadSha(ctx, version, osShaIdWin32),
	}
	squashSpec := SquashSpec{
		Version:  version,
		BaseName: extensionBaseName,
		Binaries: bins,
	}
	mustGenerateConfigFile(ctx, squashSpec)
}

func MustDownloadSha(ctx context.Context, version, osId string) string {
	url := fmt.Sprintf("https://github.com/solo-io/squash/releases/download/%v/squashctl-%v.sha256", version, osId)
	str := ""
	if err := getFileContent(url, &str); err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	}
	return str
}

func getFileContent(url string, str *string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	*str = buf.String()
	return nil
}
