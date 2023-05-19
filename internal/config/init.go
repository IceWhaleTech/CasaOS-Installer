package config

import (
	"log"

	"gopkg.in/ini.v1"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
)

type CommonModel struct {
	RuntimePath string
}

type APPModel struct {
	LogPath     string
	LogSaveName string
	LogFileExt  string
}

const (
	InstallerConfigFilePath = "/etc/casaos/installer.conf"
)

var (
	CommonInfo = &CommonModel{
		RuntimePath: "/var/run/casaos",
	}

	AppInfo = &APPModel{
		LogPath:     "/var/log/casaos",
		LogSaveName: common.InstallerServiceName,
		LogFileExt:  "log",
	}

	Cfg            *ini.File
	ConfigFilePath string
)

func InitSetup(config string) {
	ConfigFilePath = InstallerConfigFilePath
	if len(config) > 0 {
		ConfigFilePath = config
	}

	var err error

	Cfg, err = ini.LoadSources(ini.LoadOptions{Insensitive: true, AllowShadows: true}, ConfigFilePath)
	if err != nil {
		panic(err)
	}

	mapTo("common", CommonInfo)
	mapTo("app", AppInfo)
}

func mapTo(section string, v interface{}) {
	err := Cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
