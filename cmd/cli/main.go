package main

import (
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

	tag := "main"
	if *tagFlag != "" {
		tag = *tagFlag
	}

	_logger.Info("Getting release information...")

	release, err := service.GetRelease(tag)
	if err != nil {
		_logger.Error("Failed to get release: %s", err.Error())
		os.Exit(1)
	}

	if release == nil {
		_logger.Error("Release is nil")
		os.Exit(1)
	}

	_logger.Info("Release version: %s", release.Version)
}
