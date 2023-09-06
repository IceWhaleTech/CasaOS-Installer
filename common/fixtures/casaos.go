package fixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
)

func createFolderIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
}

func getReleaseYamlContent(versionTag string) string {
	return `version: ` + versionTag + `
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
	releaseContent := getReleaseYamlContent(versionTag)

	createFolderIfNotExist(filepath.Join(sysRoot, "etc", "casaos"))
	os.WriteFile(filepath.Join(sysRoot, service.CurrentReleaseLocalPath), []byte(releaseContent), 0755)
}

func SetLocalTargetRelease(sysRoot string, versionTag string) {
	releaseContent := getReleaseYamlContent(versionTag)

	createFolderIfNotExist(filepath.Join(sysRoot, "etc", "casaos"))
	os.WriteFile(filepath.Join(sysRoot, service.TargetReleaseLocalPath), []byte(releaseContent), 0755)
}

// func CacheRelease0441(cacheDir string) error {
// 	originCachePath := config.ServerInfo.CachePath
// 	config.ServerInfo.CachePath = filepath.Join("/tmp", "casaos-cache", "var", "cache", "casaos")
// 	ctx := context.Background()
// 	release, err := service.GetRelease(ctx, "unit-test-release-0.4.4-1")
// 	if err != nil {
// 		return err
// 	}

// 	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
// 	if err != nil {
// 		return err
// 	}

// 	err = service.ExtractReleasePackages(releaseFilePath, *release)
// 	if err != nil {
// 		return err
// 	}

// 	// extract module packages
// 	err = service.ExtractReleasePackages(releaseFilePath+"/linux*", *release)
// 	if err != nil {
// 		return err
// 	}

// 	// copy release file to cache path
// 	err = cp.Copy(config.ServerInfo.CachePath, cacheDir)
// 	if err != nil {
// 		return err
// 	}

// 	config.ServerInfo.CachePath = originCachePath
// 	return nil
// }

func SetZimaOS(sysRoot string) error {
	// write  sysRoot/etc/os-release file
	osReleaseContent := `PRETTY_NAME="Ubuntu 22.04.2 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.2 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
MODEL=Zima
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy`
	filePath := filepath.Join(sysRoot, "etc", "os-release")
	fmt.Println(filePath)
	os.MkdirAll(filepath.Join(sysRoot, "etc"), 0o755)

	err := os.WriteFile(filePath, []byte(osReleaseContent), 0755)
	return err
}

func SetCasaOS(sysRoot string) {

}

func SetOfflineRAUC(sysRoot string, RAUCOfflinePath string, RAUCOfflineRAUCFile string) {
	ctx := context.Background()
	internal.DownloadAs(ctx, filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile), "https://github.com/raller1028/test_rauc/releases/download/v0.4.8_offline/rauc.tar")
	fmt.Println(filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile))
	if _, err := os.Stat(filepath.Join(sysRoot, RAUCOfflinePath, RAUCOfflineRAUCFile)); os.IsNotExist(err) {
		panic("not found offline install package")
	} else {
		fmt.Println("found offline install package")
	}

}
