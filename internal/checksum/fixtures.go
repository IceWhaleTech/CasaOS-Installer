package checksum

import (
	"fmt"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

func AlwaysFail(release codegen.Release) (string, error) {
	return "", fmt.Errorf("download fail")
}
