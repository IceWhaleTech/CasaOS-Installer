package internal

import (
	"encoding/json"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/http"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

func GetReleaseFrom(releaseURL string) (*codegen.Release, error) {
	// download content from releaseURL
	response, err := http.Get(releaseURL, 30*time.Second)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// parse release
	var release *codegen.Release
	if err := json.NewDecoder(response.Body).Decode(release); err != nil {
		return nil, err
	}

	return release, nil
}
