package config

import (
	"fmt"
	"log"
	"path/filepath"

	"gopkg.in/ini.v1"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
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

type ServerModel struct {
	Mirrors   []string `ini:"mirrors,,allowshadow"`
	CachePath string
}

const InstallerConfigFilePath = "/etc/casaos/installer.conf"

var (
	CommonInfo = &CommonModel{
		RuntimePath: "/var/run/casaos",
	}

	AppInfo = &APPModel{
		LogPath:     "/var/log/casaos",
		LogSaveName: common.InstallerServiceName,
		LogFileExt:  "log",
	}

	ServerInfo = &ServerModel{
		CachePath: "/var/lib/casaos_data/rauc",
		Mirrors: []string{
			"https://casaos.oss-cn-shanghai.aliyuncs.com/IceWhaleTech/zimaos-rauc/",
			"https://raw.githubusercontent.com/IceWhaleTech/zimaos-rauc/main/",
		},
	}

	Cfg            *ini.File
	ConfigFilePath string
)

const (
	// RAUC_OFFLINE_PATH             = "/DATA/rauc/"
	RAUC_OFFLINE_PATH = "/var/lib/casaos_data/rauc/offline/"

	RAUC_OFFLINE_RELEASE_FILENAME = "release.yaml"
	OFFLINE_RAUC_TEMP_PATH        = "/tmp/offline_rauc"
)

var (
	SysRoot = "/"
	// The file name of the rauc package
	// the name can be changed by file change
	RAUC_OFFLINE_RAUC_FILENAME = "rauc.tar"
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
	mapTo("server", ServerInfo)
}

func mapTo(section string, v interface{}) {
	err := Cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}

func ReleaseDir(release codegen.Release) (string, error) {
	if release.Version == "" {
		return "", fmt.Errorf("release version is empty")
	}

	return filepath.Join(ServerInfo.CachePath, "releases", release.Version), nil
}
