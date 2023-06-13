package internal

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	httputil "github.com/IceWhaleTech/CasaOS-Common/utils/http"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"gopkg.in/yaml.v3"
)

func GetReleaseFrom(releaseURL string) (*codegen.Release, error) {
	// download content from releaseURL
	response, err := httputil.Get(releaseURL, 30*time.Second)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get release from %s - %s", releaseURL, response.Status)
	}

	// parse release
	var release codegen.Release

	if err := yaml.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func GetChecksum(filepath string) (map[string]string, error) {
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

func GetChecksumURL(release codegen.Release, mirror string) string {
	return strings.TrimSuffix(mirror, "/") + release.Checksum
}
