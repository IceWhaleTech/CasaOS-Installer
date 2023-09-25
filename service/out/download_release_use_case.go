package out

import (
	"context"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type DownloadReleaseUseCase func(ctx context.Context, release codegen.Release, checksumHandler CheckSumReleaseUseCase) (string, error)
