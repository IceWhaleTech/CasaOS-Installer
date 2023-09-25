package out

import (
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type CheckSumReleaseUseCase func(release codegen.Release) (string, error)
