package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
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
	config.ServerInfo.Mirrors = service.ChannelData[service.PublicTestChannelType]
	Save()
}

func PrivateTestChannel() {
	fmt.Println("Private Test Channel")
	config.ServerInfo.Mirrors = service.ChannelData[service.PrivateTestChannelType]
	Save()
}

func StableChannel() {
	fmt.Println("Stable Channel")
	config.ServerInfo.Mirrors = service.ChannelData[service.StableChannelType]
	Save()
}

func TestChannel() {
	fmt.Println("Test Verify Channel")
	config.ServerInfo.Mirrors = service.ChannelData[service.TestVerifyChannelType]
	Save()
}

func DisableChannel() {
	fmt.Println("Disable Channel")
	config.ServerInfo.Mirrors = service.ChannelData[service.DisableChannelType]
	Save()
}
