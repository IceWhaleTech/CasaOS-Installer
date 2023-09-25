package service

import (
	"context"
	"time"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
)

type TestService struct {
	InstallRAUCHandler func(raucPath string) error
	downloaded         bool
}

func (r *TestService) Install(release codegen.Release, sysRoot string) error {
	return nil
}

func (r *TestService) GetRelease(ctx context.Context, tag string) (*codegen.Release, error) {
	time.Sleep(2 * time.Second)
	r.downloaded = false
	return &codegen.Release{
		Version: "v0.4.8",
	}, nil
}

func (r *TestService) VerifyRelease(release codegen.Release) (string, error) {
	return "", nil
}

func (r *TestService) DownloadRelease(ctx context.Context, release codegen.Release, force bool) (string, error) {
	time.Sleep(2 * time.Second)
	r.downloaded = true
	return "", nil
}

func (r *TestService) ExtractRelease(packageFilepath string, release codegen.Release) error {
	return nil
}

func (r *TestService) GetMigrationInfo(ctx context.Context, release codegen.Release) error {
	return nil
}

func (r *TestService) DownloadAllMigrationTools(ctx context.Context, release codegen.Release) error {
	return nil
}

func (r *TestService) ShouldUpgrade(release codegen.Release, sysRoot string) bool {
	return true
}

func (r *TestService) IsUpgradable(release codegen.Release, sysRootPath string) bool {
	return r.ShouldUpgrade(release, sysRootPath) && r.downloaded
}

func (r *TestService) Launch(sysRoot string) error {
	return nil
}

func (r *TestService) PostInstall(release codegen.Release, sysRoot string) error {
	return nil
}

func (r *TestService) PostMigration(sysRoot string) error {
	return nil
}
