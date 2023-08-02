package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/samber/lo"

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

func Download(ctx context.Context, outDir, url string) (string, error) {
	filename := filepath.Base(url)
	_filepath := filepath.Join(outDir, filename)

	return _filepath, DownloadAs(ctx, _filepath, url)
}

func DownloadAs(ctx context.Context, filepath, url string) error {
	url = url + "?archive=false" // disable automatic archive extraction

	// download package
	client := getter.Client{
		Ctx:   ctx,
		Dst:   filepath,
		Mode:  getter.ClientModeFile,
		Src:   url,
		Umask: 0o022,
		Options: []getter.ClientOption{
			getter.WithProgress(NewTracker(
				func(downladed, totalSize int64) {
					// TODO: send progress event to message bus if it exists
					// logger.Info("Downloading package", zap.String("url", url), zap.Int64("downloaded", downladed), zap.Int64("totalSize", totalSize))
				},
			)),
		},
	}

	return client.Get()
}

func Extract(tarFilePath, destinationFolder string) error {
	// to check tarFilePath is a tar.gz file and destinationFolder is a folder
	if strings.HasSuffix(tarFilePath, ".tar.gz") {
		exec.Command("tar", "-xzf", tarFilePath, "-C", destinationFolder).Run()
	}
	return nil
}

// extract each archive in dir
func BulkExtract(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		return Extract(path, dir)
	})
}

func InstallRelease(releaseDir string, sysrootPath string) error {
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

func InstallDependencies() error {
	CASA_DEPANDS_PACKAGE := []string{"curl", "smartmontools", "parted", "ntfs-3g", "net-tools", "udevil", "samba", "cifs-utils", "rclone"}
	Output, err := exec.Command("apt", "update").Output()
	fmt.Println(string(Output))
	if err != nil {
		return err
	}

	for _, depName := range CASA_DEPANDS_PACKAGE {
		Output, err := exec.Command("apt", "install", "-y", depName).Output()
		fmt.Println(string(Output))
		if err != nil {
			return err
		}
	}
	return nil
}

func getDockerRunningStatus() bool {
	//systemctl is-active docker
	_, writer, err := os.Pipe()
	if err != nil {
		return false
	}
	cmd := exec.Command("systemctl", "is-active", "docker")

	cmd.Stdout = writer

	if cmd.String() == "active" {
		return true
	} else {
		return false
	}
}

func isDockerInstalled() bool {
	// command -v docker
	cmd := exec.Command("command", "-v", "docker")
	if cmd == nil {
		return true
	} else {
		return false
	}

}

func installDocker() error {
	// to install docker
	// curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
	exec.Command("curl", "-fsSL", "https://get.docker.com", "|", "bash", "-s", "docker", "--mirror", "Aliyun")

	return nil
}
func InstallDocker() error {
	// to check docker is running
	isDockerRunning := getDockerRunningStatus()
	if isDockerRunning {
		logger.Info("docker is running. skip install docker")
		return nil
	}

	// to check docker is installed
	if isDockerInstalled() {
		return nil
	}

	// to install docker
	err := installDocker()
	if err != nil {
		return err
	}
	// to run docker

	isDockerRunning = getDockerRunningStatus()
	if !isDockerRunning {
		return fmt.Errorf("docker is launch failed")
	}
	return nil
}
