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

func GetPackageURLByCurrentArch(release codegen.Release, mirror string) (string, error) {
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
			return strings.TrimSuffix(mirror, "/") + pkg.Path, nil
		}
	}

	return "", fmt.Errorf("package not found for architecture: %s", arch)
}

func DownloadAndExtractPackage(ctx context.Context, outDir, packageURL string) error {
	// download package
	client := &getter.Client{
		Ctx:   ctx,
		Dst:   outDir,
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
		return err
	}

	// extract each archive in releaseDir
	if err := filepath.WalkDir(outDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		decompressor := NewDecompressor(path)
		return decompressor.Decompress(outDir, path, true, 0o022)
	}); err != nil {
		return err
	}

	return nil
}

func InstallRelease(ctx context.Context, releaseDir string, sysrootPath string) error {
	srcSysroot := filepath.Join(releaseDir, "build", "sysroot") + "/"
	if _, err := os.Stat(srcSysroot); err != nil {
		return err
	}

	return file.CopyDir(srcSysroot, sysrootPath, "")
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
