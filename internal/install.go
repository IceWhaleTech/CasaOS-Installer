package internal

import (
	"fmt"
	"runtime"

	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

func InstallRelease(release codegen.Release) error {
	// get current arch
	arch := runtime.GOARCH

	if arch == "arm" {
		arch = "arm-7"
	}

	if !lo.Contains([]string{string(codegen.Amd64), string(codegen.Arm64), string(codegen.Arm7)}, arch) {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	var packageURL string
	for _, pkg := range release.Packages {
		if string(pkg.Architecture) == arch {
			packageURL = pkg.URL
			break
		}
	}

	// prepare workdir

	// download package
	logger.Info("Downloading package", zap.String("url", packageURL))

	panic("not implemented")
}
