package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
)

type TarService struct {
}

func (r *TarService) Install(release codegen.Release, sysRoot string) error {
	return nil
}

func (r *TarService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	return FetchRelease(ctx, tag, GitHubBranchTagReleaseUrl)
}

func DownloadUninstallScript(ctx context.Context, sysRoot string) (string, error) {
	CASA_UNINSTALL_URL := "https://get.casaos.io/uninstall/v0.4.0"
	CASA_UNINSTALL_PATH := filepath.Join(sysRoot, "/usr/bin/casaos-uninstall")
	// to delete the old uninstall script when the script is exsit
	if _, err := os.Stat(CASA_UNINSTALL_PATH); err == nil {
		// 删除文件
		err := os.Remove(CASA_UNINSTALL_PATH)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Old uninstall script deleted successfully")
		}
	}

	// to download the new uninstall script
	if err := internal.DownloadAs(ctx, CASA_UNINSTALL_PATH, CASA_UNINSTALL_URL); err != nil {
		return CASA_UNINSTALL_PATH, err
	}
	// change the permission of the uninstall script
	if err := os.Chmod(CASA_UNINSTALL_PATH, 0o755); err != nil {
		return CASA_UNINSTALL_PATH, err
	}

	return "", nil
}

func VerifyUninstallScript(sysRoot string) bool {
	// to check the present of file
	// how to do the test? the uninstall is always in the same place?
	return !file.CheckNotExist(filepath.Join(sysRoot, "/usr/bin/casaos-uninstall"))
}
