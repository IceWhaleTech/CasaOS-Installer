package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
)

var (
	_logger *Logger
	// _status *version.GlobalMigrationStatus

	commit = "private build"
	date   = "private build"
)

func main() {
	debugFlag := flag.Bool("d", true, "debug")
	versionFlag := flag.Bool("v", false, "version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", common.InstallerVersion)
		os.Exit(0)
	}

	fmt.Println("git commit:", commit)
	fmt.Println("build date:", date)

	_logger = NewLogger()

	if *debugFlag {
		_logger.DebugMode = true
	}
}
