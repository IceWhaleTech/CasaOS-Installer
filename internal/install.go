package internal

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/go-getter"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

func GetPackageURLByCurrentArch(release codegen.Release) (string, error) {
	// get current arch
	arch := runtime.GOARCH

	if arch == "arm" {
		arch = "arm-7"
	}

	if !lo.Contains([]string{string(codegen.Amd64), string(codegen.Arm64), string(codegen.Arm7)}, arch) {
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	for _, pkg := range release.Packages {
		if string(pkg.Architecture) == arch {
			return pkg.URL, nil
		}
	}

	return "", fmt.Errorf("package not found for architecture: %s", arch)
}

func DownloadPackage(ctx context.Context, packageURL string) (string, error) {
	// prepare workdir
	tempDir, err := os.MkdirTemp("", "casaos-installer-*")
	if err != nil {
		return "", err
	}

	// download package
	client := &getter.Client{
		Ctx:  ctx,
		Src:  packageURL,
		Dst:  tempDir,
		Mode: getter.ClientModeDir,
		Options: []getter.ClientOption{
			getter.WithProgress(NewTracker(
				func(downladed, totalSize int64) {
					// TODO: send progress event to message bus if it exists
					logger.Info("Downloading package", zap.String("url", packageURL), zap.Int64("downloaded", downladed), zap.Int64("totalSize", totalSize))
				},
			)),
		},
	}

	if err := client.Get(); err != nil {
		if err := os.RemoveAll(tempDir); err != nil {
			logger.Error("error while removing temp dir", zap.Error(err), zap.String("dir", tempDir))
		}
		return "", err
	}

	return tempDir, nil
}

func InstallRelease(ctx context.Context, release codegen.Release) error {
	packageURL, err := GetPackageURLByCurrentArch(release)
	if err != nil {
		return err
	}

	// prepare workdir
	tempDir, err := DownloadPackage(ctx, packageURL)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	panic("not implemented")
}
