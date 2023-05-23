package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httputil "github.com/IceWhaleTech/CasaOS-Common/utils/http"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
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
	var release *codegen.Release
	if err := json.NewDecoder(response.Body).Decode(release); err != nil {
		return nil, err
	}

	return release, nil
}
