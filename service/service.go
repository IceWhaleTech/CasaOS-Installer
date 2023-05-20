package service

var Installer *InstallerService

func Initialize() {
	Installer = NewInstallerService()
}
