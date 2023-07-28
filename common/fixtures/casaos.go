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
