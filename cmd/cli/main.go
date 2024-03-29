package main

import (
	"github.com/gookit/color"
)

var (
	_logger *Logger

	commit  = "private build"
	date    = "private build"
	sysRoot = "/"
)

type InternalLogWriter struct {
	Color color.Color
}

func (l InternalLogWriter) Write(p []byte) (n int, err error) {
	l.Color.Print(string(p))
	return len(p), nil
}

func main() {
	// 	tagFlag := flag.String("t", "", "tag")
	// 	versionFlag := flag.Bool("v", false, "version")
	// 	downloadOnlyFlag := flag.Bool("d", false, "download only")
	// 	flag.Parse()

	// 	if *versionFlag {
	// 		fmt.Printf("v%s\n", common.InstallerVersion)
	// 		fmt.Println()
	// 		fmt.Println("git commit:", commit)
	// 		fmt.Println("build date:", date)
	// 		os.Exit(0)
	// 	}

	// 	{
	// 		// CLI logger
	// 		_logger = NewLogger()

	// 		// internal logger
	// 		logger.LogInitWithWriterSyncers(zapcore.AddSync(InternalLogWriter{Color: color.FgDarkGray}))
	// 	}

	// 	if os.Getuid() != 0 {
	// 		_logger.Info("Root privileges are required to run this program.")
	// 		os.Exit(1)
	// 	}

	// 	tag := service.GetReleaseBranch(sysRoot)

	// 	if tagFlag != nil && *tagFlag != "" {
	// 		tag = *tagFlag
	// 	}

	// 	downloadOnly := false
	// 	if downloadOnlyFlag != nil {
	// 		downloadOnly = *downloadOnlyFlag
	// 	}

	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	defer cancel()

	// 	// get release information
	// 	_logger.Info("🟨 Getting release information...")

	// 	release, err := service.GetRelease(ctx, tag)
	// 	if err != nil {
	// 		_logger.Error("🟥 Failed to get release: %s", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	if release == nil {
	// 		_logger.Error("🟥 Release is nil")
	// 		os.Exit(1)
	// 	}

	// 	_logger.Info("🟩 Release found: %s", release.Version)

	// 	// stop old module
	// 	_logger.Info("🟨 Stopping old module...")
	// 	if err = service.StopModule(*release); err != nil {
	// 		_logger.Error("🟥 Failed to stop old module: %s", err.Error())
	// 	}
	// 	_logger.Info("🟩 Stopping old module: %s", release.Version)

	// 	// install dep
	// 	_logger.Info("🟨 Install dependencies...")
	// 	err = internal.InstallDocker()
	// 	if err != nil {
	// 		_logger.Error("🟥 Failed to install docker: %s", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	err = service.InstallDependencies(*release, sysRoot)
	// 	if err != nil {
	// 		_logger.Error("🟥 Failed to install dependencies: %s", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	_logger.Info("🟩 Dependencies Installed.")

	// 	// download release
	// 	_logger.Info("🟨 Downloading release %s...", release.Version)
	// 	releaseFilePath, err := service.DownloadRelease(ctx, *release, false)
	// 	if err != nil {
	// 		_logger.Error("Failed to download release: %s", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	_logger.Info("🟩 Release downloaded: %s", releaseFilePath)

	// 	// verify release
	// 	_logger.Info("🟨 Verifying release...")
	// 	if _, err := service.VerifyRelease(*release); err != nil {
	// 		_logger.Error("🟥 Release verification failed: %s", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	_logger.Info("🟩 Release verified.")

	// 	_logger.Info("🟨 Installing release...")
	// 	// if err := service.InstallCasaOSPackages(*release, sysRoot); err != nil {
	// 	// 	_logger.Error("🟥 Failed to install release: %s", err.Error())
	// 	// 	os.Exit(1)
	// 	// }
	// 	// _logger.Info("🟩 Release installed.")

	// 	// post install release
	// 	_logger.Info("🟨 Handle Post Release Install ...")
	// 	if err := service.PostReleaseInstall(*release, sysRoot); err != nil {
	// 		_logger.Error("🟥 Failed to Handle Post Release Install: %s", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	_logger.Info("🟩 Handle Post Release Install completed")

	// 	_logger.Info("🟨 Downloading migration tools...")
	// 	downloaded, err := service.DownloadAllMigrationTools(ctx, *release, sysRoot)
	// 	if err != nil {
	// 		_logger.Error("🟥 Failed to download migration tools: %s", err.Error())
	// 		os.Exit(1)
	// 	}

	// 	if downloaded {
	// 		_logger.Info("🟩 Migration tools downloaded.")

	// 		_logger.Info("🟨 Verifying migration tools...")
	// 		if !service.VerifyAllMigrationTools(*release, sysRoot) {
	// 			_logger.Error("🟥 Migration tools verification failed: %s", err.Error())
	// 			os.Exit(1)
	// 		}
	// 		_logger.Info("🟩 Migration tools verified.")
	// 	} else {
	// 		_logger.Info("🟩 No migration tools to download.")
	// 	}

	// 	if downloadOnly {
	// 		_logger.Info("🟩 Download complete.")
	// 		os.Exit(0)
	// 	}

	// 	isCasaOS := true
	// 	if isCasaOS {
	// 		_logger.Info("🟨 Start Migration...")

	// 		if err := service.StartMigration(sysRoot); err != nil {
	// 			_logger.Error("🟥 Failed to Migration: %s", err.Error())
	// 			os.Exit(1)
	// 		}
	// 		_logger.Info("🟩 Migration complete.")
	// 	}

	// 	_logger.Info("🟨 Installing modules...")
	// 	if err := service.ExecuteModuleInstallScript(releaseFilePath, *release); err != nil {
	// 		_logger.Error("🟥 Failed to install modules: %s", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	_logger.Info("🟩 Modules installed.")

	// 	_logger.Info("🟨 Enable services...")
	// 	if err := service.LaunchModule(*release); err != nil {
	// 		_logger.Error("🟥 Failed to enable services: %s", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	_logger.Info("🟩 Services enabled.")
	// 	// download uninstall script
	// 	_logger.Info("🟨 Downloading uninstall script ...")
	// 	if _, err = service.DownloadUninstallScript(ctx, sysRoot); err != nil {
	// 		_logger.Error("Downloading uninstall script: %s", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	_logger.Info("🟩 Uninstall script Downloaded")

	//	if service.VerifyUninstallScript(sysRoot) {
	//		_logger.Info("🟨 uninstall script is installed")
	//	} else {
	//
	//		panic("🟥 uninstall script is not installed")
	//	}
}
