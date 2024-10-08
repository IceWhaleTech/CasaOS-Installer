package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
)

func main() {
	config.InitSetup(config.InstallerConfigFilePath)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("unknow options")
		Help()
		os.Exit(0)
	}

	switch args[1] {
	case "public":
		PublicTestChannel()
	case "private":
		PrivateTestChannel()
	case "stable":
		StableChannel()
	case "internal":
		TestChannel()
	case "disable":
		DisableChannel()
	case "-h":
		Help()
	default:
		fmt.Println("unknow options")
		Help()
	}
}

func Help() {
	fmt.Println("Usage: channel <command>")
	fmt.Println("Commands:")
	fmt.Println("  public     Set channel to public test")
	fmt.Println("  stable     Set channel to stable")
	fmt.Println("  -h         Show help")
}

func Save() {
	config.Cfg.Section("server").Key("mirrors").SetValue(strings.Join(config.ServerInfo.Mirrors, ","))
	err := config.Cfg.SaveTo(config.ConfigFilePath)
	if err != nil {
		fmt.Printf("Fail to save file: %v", err)
	}
	// restart casaos installer
	err = exec.Command("systemctl", "restart", "casaos-installer").Run()
	if err != nil {
		fmt.Printf("Fail to restart casaos-installer: %v", err)
	}
}

func PublicTestChannel() {
	fmt.Println("Public Test Channel")
	config.ServerInfo.Mirrors = []string{
		"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/public-test/",
		"https://raw.githubusercontent.com/IceWhaleTech/ZimaOS/refs/heads/main/public-test/",
	}
	Save()
}

func PrivateTestChannel() {
	fmt.Println("Private Test Channel")
	config.ServerInfo.Mirrors = []string{"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/private-test/"}
	Save()
}

func StableChannel() {
	fmt.Println("Stable Channel")
	config.ServerInfo.Mirrors = []string{
		"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/",
		"https://raw.githubusercontent.com/IceWhaleTech/ZimaOS/refs/heads/main/",
	}
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
