package fixtures

import (
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
)

const rauc_info_048 = `Compatible: 	'zimaos-zimacube'
Version:    	'0.5.0.4'
Description:	'dmVyc2lvbjogdjAuNS4wLjQKcmVsZWFzZV9ub3RlczogfAogICMgcHJpdmF0ZSB0ZXN0Cm1pcnJvcnM6CiAgLSBodHRwczovL2Nhc2Fvcy5vc3MtY24tc2hhbmdoYWkuYWxpeXVuY3MuY29tL0ljZVdoYWxlVGVjaApwYWNrYWdlczoKICAtIHBhdGg6IC96aW1hb3MtcmF1Yy9yZWxlYXNlcy9kb3dubG9hZC90ZXN0L3ppbWFvc196aW1hY3ViZS0wLjQuOC5yYXVjYgogICAgYXJjaGl0ZWN0dXJlOiBhbWQ2NApjaGVja3N1bXM6IC9nZXQvcmVsZWFzZXMvZG93bmxvYWQvdjAuNC40LTEvY2hlY2tzdW1zLnR4dAptb2R1bGVzOgogIC0gbmFtZTogY2FzYW9zLWdhdGV3YXkKICAgIHNob3J0OiBnYXRld2F5CiAgLSBuYW1lOiBjYXNhb3MtdXNlci1zZXJ2aWNlCiAgICBzaG9ydDogdXNlci1zZXJ2aWNlCiAgLSBuYW1lOiBjYXNhb3MtbWVzc2FnZS1idXMKICAgIHNob3J0OiBtZXNzYWdlLWJ1cwogIC0gbmFtZTogY2FzYW9zCiAgICBzaG9ydDogY2FzYW9zCiAgLSBuYW1lOiBjYXNhb3MtbG9jYWwtc3RvcmFnZQogICAgc2hvcnQ6IGxvY2FsLXN0b3JhZ2UKICAtIG5hbWU6IGNhc2Fvcy1hcHAtbWFuYWdlbWVudAogICAgc2hvcnQ6IGFwcC1tYW5hZ2VtZW50'
Build:      	'(null)'
Hooks:      	'install-check'
Bundle Format: 	plain

3 Images:
  [boot]
	Filename:  boot.vfat
	Checksum:  055cbef657f1dcf1995124604994bc6d2d6477c47404f4bc11b6050666a4ffaa
	Size:      33554432
	Hooks:     install
  [kernel]
	Filename:  kernel.img
	Checksum:  b39752abf9380e9e22c7e76bba898782056616c856823ab1a237cb9ae7e4b29b
	Size:      14413824
	Hooks:     post-install
  [rootfs]
	Filename:  rootfs.img
	Checksum:  390166cd2c16b0c8389ac814b01a5623b3137aea455c58f2727d5623cbd67b75
	Size:      490135552
	Hooks:

Certificate Chain:
 0 Subject: O = IceWhale Technology, CN = IceWhale Technology Development-1
   Issuer: O = IceWhale Technology, CN = IceWhale Technology OTA Development
   SPKI sha256: 96:A9:8A:2D:12:E3:6F:DE:ED:B1:0B:C8:26:2D:7C:EA:30:34:B5:15:1E:E6:AB:7C:DA:AD:F9:DC:DC:84:01:AD
   Not Before: Jan  1 00:00:00 1970 GMT
   Not After:  Dec 31 23:59:59 9999 GMT
 1 Subject: O = IceWhale Technology, CN = IceWhale Technology OTA Development
   Issuer: O = IceWhale Technology, CN = IceWhale Technology OTA Development
   SPKI sha256: FE:BE:C2:D0:42:16:92:F4:85:5D:3D:71:C9:79:FC:D9:16:AE:9E:73:EB:74:48:4D:3E:D2:23:54:AF:D5:05:B4
   Not Before: Jan  1 00:00:00 1970 GMT
   Not After:  Dec 31 23:59:59 9999 GMT`

func SetOfflineRAUCMock_0504(sysRoot string) error {
	service.MockContent = rauc_info_048
	return nil
}

func SetOfflineRAUCRelease_050(sysRoot string) error {
	service.MockContent = rauc_info_048
	release := &codegen.Release{
		Version: "v0.5.0",
	}
	err := internal.WriteReleaseToLocal(release, filepath.Join(sysRoot, config.OFFLINE_RAUC_TEMP_PATH, common.ReleaseYAMLFileName))
	return err
}
