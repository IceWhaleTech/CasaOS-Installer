package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

var (
	commit = "private build"
	date   = "private build"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// parse arguments and intialize
	{
		configFlag := flag.String("c", "", "config file path")
		versionFlag := flag.Bool("v", false, "version")

		flag.Parse()

		if *versionFlag {
			fmt.Printf("v%s\n", common.InstallerVersion)
			os.Exit(0)
		}

		println("git commit:", commit)
		println("build date:", date)

		config.InitSetup(*configFlag)
	}
}
