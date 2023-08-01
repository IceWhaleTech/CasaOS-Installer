package fixtures

import (
	"os"
	"path/filepath"
)

func createFolderIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
}

func SetCasaOS043(sysRoot string, module string) {
	casaos043VersionScript := "#! /usr/bin/python3\nprint(\"v0.4.3\")"
	createFolderIfNotExist(filepath.Join(sysRoot, "/usr", "bin"))
	casaosPath := filepath.Join(sysRoot, "/usr", "bin", module)
	os.WriteFile(casaosPath, []byte(casaos043VersionScript), 0755)
}

func SetCasaOS035(sysRoot string, module string) {
	casaos043VersionScript := "#! /usr/bin/python3\nprint(\"v0.3.5\")"
	createFolderIfNotExist(filepath.Join(sysRoot, "/usr", "bin"))
	casaosPath := filepath.Join(sysRoot, "/usr", "bin", module)
	os.WriteFile(casaosPath, []byte(casaos043VersionScript), 0755)
}

func SetCasaOSVersion(sysRoot string, module string, versionTag string) {
	casaos043VersionScript := "#! /usr/bin/python3\nprint(\"" + versionTag + "\")"
	createFolderIfNotExist(filepath.Join(sysRoot, "/usr", "bin"))
	casaosPath := filepath.Join(sysRoot, "/usr", "bin", module)
	os.WriteFile(casaosPath, []byte(casaos043VersionScript), 0755)
}

func SetLocalRelease(sysRoot string, versionTag string) {
	releaseContent := `version: ` + versionTag + `
release_notes: |
  TOOD: add release notes
mirrors:
  - http://casaos.io/does/not/exist/test
  - https://github.com/IceWhaleTech
packages:
  - path: /get/releases/download/v0.4.4-alpha2/casaos-amd64-v0.4.4-alpha2.tar.gz
    architecture: amd64
  - path: /get/releases/download/v0.4.4-alpha2/casaos-arm64-v0.4.4-alpha2.tar.gz
    architecture: arm64
  - path: /get/releases/download/v0.4.4-alpha2/casaos-arm-7-v0.4.4-alpha2.tar.gz
    architecture: armv7
checksums: /get/releases/download/v0.4.4-alpha2/checksums.txt
modules:
  - name: casaos-gateway
    short: gateway
  - name: casaos-user-service
    short: user-service
  - name: casaos
    short: casaos
  - name: casaos-local-storage
    short: local-storage
  - name: casaos-message-bus
    short: message-bus
  - name: casaos-app-management
    short: app-management
`

	createFolderIfNotExist(filepath.Join(sysRoot, "etc", "casaos"))
	os.WriteFile(filepath.Join(sysRoot, "etc", "casaos", "release.yaml"), []byte(releaseContent), 0755)
}
