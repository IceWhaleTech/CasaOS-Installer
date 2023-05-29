package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
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

func DownloadAndExtractPackage(ctx context.Context, packageURL string) (string, error) {
	// prepare workdir
	tempDir, err := os.MkdirTemp("", "casaos-installer-*")
	if err != nil {
		return "", err
	}

	// download package
	client := &getter.Client{
		Ctx:   ctx,
		Dst:   tempDir,
		Mode:  getter.ClientModeDir,
		Src:   packageURL,
		Umask: 0x022,
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

func InstallRelease(ctx context.Context, release codegen.Release, sysroot string) error {
	packageURL, err := GetPackageURLByCurrentArch(release)
	if err != nil {
		return err
	}

	// prepare workdir
	tempDir, err := DownloadAndExtractPackage(ctx, packageURL)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	if err := filepath.WalkDir(tempDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		decompressor := NewDecompressor(path)
		return decompressor.Decompress(tempDir, path, true, 0o022)
	}); err != nil {
		return err
	}

	srcSysroot := filepath.Join(tempDir, "build", "sysroot") + "/"
	if _, err := os.Stat(srcSysroot); err != nil {
		return err
	}

	return file.CopyDir(srcSysroot, sysroot, "")
}

func NewDecompressor(filepath string) getter.Decompressor {
	matchingLen := 0
	archiveV := ""
	for k := range getter.Decompressors {
		if strings.HasSuffix(filepath, "."+k) && len(k) > matchingLen {
			archiveV = k
			matchingLen = len(k)
		}
	}
	return getter.Decompressors[archiveV]
}
