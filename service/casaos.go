package service

import "github.com/IceWhaleTech/CasaOS-Installer/codegen"

// dependent config.ServerInfo.CachePath
func InstallCasaOSPackages(release codegen.Release, sysRoot string) error {
	releaseFilePath, err := VerifyRelease(release)
	if err != nil {
		return err
	}

	// extract packages
	err = ExtractReleasePackages(releaseFilePath, release)
	if err != nil {
		return err
	}

	// extract module packages
	err = ExtractReleasePackages(releaseFilePath+"/linux*", release)
	if err != nil {
		return err
	}

	err = InstallRelease(release, sysRoot)
	if err != nil {
		return err
	}
	return nil
}
