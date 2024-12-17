package service

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/holoplot/go-rauc/rauc"
	"go.uber.org/zap"
)

const (
	FlagUpgradeFile = "/var/lib/casaos/upgradeInfo.txt"
)

func ExtractRAUCRelease(packageFilepath string, release codegen.Release) error {
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return err
	}

	return internal.BulkExtract(releaseDir)
}

// dependent config.ServerInfo.CachePath
func InstallRAUC(release codegen.Release, sysRoot string, InstallRAUCHandler func(raucPath string) error) error {
	// to check rauc tar

	raucFilePath, err := RAUCFilePath(release)
	if err != nil {
		return err
	}

	err = InstallRAUCHandler(raucFilePath)
	if err != nil {
		log.Println("VerifyRAUC() failed: ", err.Error())
		return err
	}

	return nil
}

func InstallRAUCImp(raucFilePath string) error {
	// install rauc
	logger.Info("installing rauc", zap.String("rauc path", raucFilePath))

	raucInstaller, err := rauc.InstallerNew()
	if err != nil {
		logger.Error("new rauc installer fail", zap.Error(err))
		return nil
	}

	compatible, version, err := raucInstaller.Info(raucFilePath)
	if err != nil {
		logger.Error("get rauc info fail", zap.Error(err))
		return err
	}
	log.Printf("Info(): compatible=%s, version=%s", compatible, version)

	err = raucInstaller.InstallBundle(raucFilePath, rauc.InstallBundleOptions{})
	if err != nil {
		logger.Error("install rauc fail", zap.Error(err))
		return err
	}

	return nil
}

func MockInstallRAUC(raucFilePath string) error {
	// to check file exist
	fmt.Println("filename: ", raucFilePath)
	if _, err := os.Stat(raucFilePath); os.IsNotExist(err) {
		return fmt.Errorf("not found rauc install package")
	}

	return nil
}

func PostInstallRAUC(release codegen.Release, sysRoot string) error {
	// the code is for user experience.
	time.Sleep(5 * time.Second)
	RebootSystem()
	return nil
}

func OfflineRAUCFilePath() string {
	return filepath.Join(config.SysRoot, config.RAUC_OFFLINE_PATH, config.RAUC_OFFLINE_RAUC_FILENAME)
}

func RAUCFilePath(release codegen.Release) (string, error) {
	// 这个是验证解压之后的包。
	releaseDir, err := config.ReleaseDir(release)
	if err != nil {
		return "", err
	}

	packageURL, err := internal.GetPackageURLByCurrentArch(release, "")
	if err != nil {
		return "", err
	}

	packageFilename := filepath.Base(packageURL)

	packageFilePath := filepath.Join(releaseDir, packageFilename)

	// packageFilePath = packageFilePath[:len(packageFilePath)-len(".tar")] + ".raucb"
	// to check file exist
	logger.Info("verify rauc whether in cache", zap.String("rauc file path", packageFilePath))
	if _, err := os.Stat(packageFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("not found rauc install package")
	}
	return packageFilePath, nil
}

func MarkGood() error {
	return exec.Command("rauc", "status", "mark-good").Run()
}

func RebootSystem() {
	exec.Command("reboot").Run()
}

func getFreeMemory() (uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return 0, fmt.Errorf("unexpected line in /proc/meminfo: %s", line)
			}
			mem, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return 0, err
			}
			// /proc/meminfo中内存的单位是KB，所以需要转换成MB
			return mem / 1024, nil
		}
	}
	if scanner.Err() != nil {
		return 0, scanner.Err()
	}
	return 0, fmt.Errorf("did not find MemAvailable in /proc/meminfo")
}

func CheckMemory() error {
	mem, err := getFreeMemory()
	if mem < 600 {
		return fmt.Errorf("memory is less than 600MB")
	}
	if err != nil {
		return err
	}
	return nil
}

var MockContent string = ``

func MockRAUCInfo(content string) (string, error) {
	return MockContent, nil
}

func GetRAUCInfo(path string) (string, error) {
	cmd := exec.Command("rauc", "info", path)
	var out bytes.Buffer
	var errReason bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errReason
	err := cmd.Run()

	if err != nil {
		logger.Error("get info from rauc fail", zap.Error(err), zap.String("reason", errReason.String()), zap.String("path", path))
		lines := strings.Split(errReason.String(), "\n")
		if len(lines) < 2 {
			return "", fmt.Errorf("get info from rauc fail: %s", errReason.String())
		}
		theLastTwoLine := lines[len(lines)-2]
		return "", fmt.Errorf("get info from rauc fail: %s", theLastTwoLine)
	}

	return out.String(), nil
}

func GetDescription(raucInfo string) (string, error) {
	lines := strings.Split(raucInfo, "\n")
	if len(lines) < 8 {
		return "", fmt.Errorf("unexpected output: less than 8 lines")
	}

	line := lines[2]
	prefix := "Description:\t'"
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("unexpected line format: %s", line)
	}

	description := strings.TrimPrefix(line, prefix)
	description = strings.TrimSuffix(description, "'")

	return description, nil
}
