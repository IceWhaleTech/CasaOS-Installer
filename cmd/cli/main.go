package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/gookit/color"
	"go.uber.org/zap/zapcore"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
)

var (
	_logger *Logger

	commit = "private build"
	date   = "private build"
)

type InternalLogWriter struct {
	Color color.Color
}

func (l InternalLogWriter) Write(p []byte) (n int, err error) {
	l.Color.Print(string(p))
	return len(p), nil
}

func main() {
	tagFlag := flag.String("t", "", "tag")
	versionFlag := flag.Bool("v", false, "version")
	downloadOnlyFlag := flag.Bool("d", false, "download only")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", common.InstallerVersion)
		fmt.Println()
		fmt.Println("git commit:", commit)
		fmt.Println("build date:", date)
		os.Exit(0)
	}

	{
		// CLI logger
		_logger = NewLogger()

		// internal logger
		logger.LogInitWithWriterSyncers(zapcore.AddSync(InternalLogWriter{Color: color.FgDarkGray}))
	}

	if os.Getuid() != 0 {
		_logger.Info("Root privileges are required to run this program.")
		os.Exit(1)
	}
	tag := "main"
	if tagFlag != nil && *tagFlag != "" {
		tag = *tagFlag
	}

	downloadOnly := false
	if downloadOnlyFlag != nil {
		downloadOnly = *downloadOnlyFlag
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_logger.Info("🟨 Getting release information...")

	release, err := service.GetRelease(ctx, tag)
	if err != nil {
		_logger.Error("🟥 Failed to get release: %s", err.Error())
		os.Exit(1)
	}

	if release == nil {
		_logger.Error("🟥 Release is nil")
		os.Exit(1)
	}

	_logger.Info("🟨 Downloading release %s...", release.Version)
	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	if err != nil {
		_logger.Error("Failed to download release: %s", err.Error())
		os.Exit(1)
	}
	_logger.Info("🟩 Release file path: %s", releaseFilePath)

	_logger.Info("🟨 Verifying release...")
	if err := service.VerifyReleaseChecksum(*release); err != nil {
		_logger.Error("🟥 Release verification failed: %s", err.Error())
		os.Exit(1)
	}

	_logger.Info("🟨 Downloading migration tools...")
	if err := service.DownloadAllMigrationTools(ctx, *release); err != nil {
		_logger.Error("🟥 Failed to download migration tools: %s", err.Error())
		os.Exit(1)
	}
	_logger.Info("🟩 Migration tools downloaded.")

	_logger.Info("🟨 Verifying migration tools...")
	if !service.VerifyAllMigrationTools(*release) {
		_logger.Error("🟥 Migration tools verification failed")
		os.Exit(1)
	}

	if downloadOnly {
		_logger.Info("🟩 Download complete.")
		os.Exit(0)
	}

	_logger.Info("🟨 Installing release...")
	if err := service.InstallRelease(ctx, *release, "/"); err != nil {
		_logger.Error("🟥 Failed to install release: %s", err.Error())
		os.Exit(1)
	}

	_logger.Info("🟩 Release installed.")

	// TODO: make sure `casaos-uninstall` script is installed
}