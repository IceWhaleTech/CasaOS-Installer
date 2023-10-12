package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

var client = resty.New()

func init() {
	client.
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second)
}

func GetReleaseFromLocal(releasePath string) (*codegen.Release, error) {
	// open the yaml file
	f, err := os.Open(releasePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// decode the yaml file
	var release codegen.Release
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func GetReleaseFromContent(content []byte) (*codegen.Release, error) {
	// decode the yaml file
	var release codegen.Release
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func GetReleaseFrom(ctx context.Context, releaseURL string) (*codegen.Release, error) {
	// download content from releaseURL
	response, err := client.R().SetContext(ctx).Get(releaseURL)
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get release from %s - %s", releaseURL, response.Status())
	}

	// parse release
	var release codegen.Release

	if err := yaml.Unmarshal(response.Body(), &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func GetChecksums(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	checksums := map[string]string{}

	// get checksums
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) != 2 {
			continue
		}

		checksums[fields[1]] = fields[0]
	}

	return checksums, nil
}

func GetChecksumsURL(release codegen.Release, mirror string) string {
	return strings.TrimSuffix(mirror, "/") + release.Checksums
}
