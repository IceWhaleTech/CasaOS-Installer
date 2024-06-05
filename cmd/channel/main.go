package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

func main() {
	config.InitSetup(config.InstallerConfigFilePath)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Please provide a command.")
		os.Exit(1)
	}

	switch args[1] {
	case "public":
		PublicTestChannel()
	case "private":
		PrivateTestChannel()
	case "stable":
		StableChannel()
	case "test":
		TestChannel()
	default:
		fmt.Println("unknow channel.")
		os.Exit(1)
	}
}

func Save() {
	config.Cfg.Section("server").Key("mirrors").SetValue(strings.Join(config.ServerInfo.Mirrors, ","))
	err := config.Cfg.SaveTo(config.ConfigFilePath)
	if err != nil {
		fmt.Printf("Fail to save file: %v", err)
	}
}

func PublicTestChannel() {
	fmt.Println("Public Test Channel")
	config.ServerInfo.Mirrors = []string{"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/public-test/"}
	Save()
}

func PrivateTestChannel() {
	fmt.Println("Private Test Channel")
	config.ServerInfo.Mirrors = []string{"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/private-test/"}
	Save()
}

func StableChannel() {
	fmt.Println("Stable Channel")
	config.ServerInfo.Mirrors = []string{"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/"}
	Save()
}

func TestChannel() {
	fmt.Println("Test Verify Channel")
	config.ServerInfo.Mirrors = []string{"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/test-verification-channel/"}
	Save()
}

func DisableChannel() {
	fmt.Println("Disable Channel")
	config.ServerInfo.Mirrors = []string{"https://localhost/IceWhaleTech/zimaos-rauc/"}
	Save()
}
