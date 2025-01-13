//go:generate bash -c "mkdir -p codegen && go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0  -generate types,server,spec -package codegen api/installer/openapi.yaml > codegen/api.go"
//go:generate bash -c "mkdir -p codegen/message_bus && go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.1.0 -generate types,client -package message_bus https://raw.githubusercontent.com/IceWhaleTech/CasaOS-MessageBus/main/api/message_bus/openapi.yaml > codegen/message_bus/api.go"

package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"

	util_http "github.com/IceWhaleTech/CasaOS-Common/utils/http"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/route"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/coreos/go-systemd/daemon"
	"go.uber.org/zap"
)

var (
	commit = "private build"
	date   = "private build"

	//go:embed api/index.html
	_docHTML string

	//go:embed api/installer/openapi.yaml
	_docYAML string
	sysRoot  = "/"

	//go:embed build/sysroot/etc/casaos/installer.conf.sample
	_confSample string
)

func init() {
	configFlag := flag.String("c", "", "config file path")
	versionFlag := flag.Bool("v", false, "version")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", common.InstallerVersion)
		os.Exit(0)
	}

	println("git commit:", commit)
	println("build date:", date)

	ConfigFilePath := filepath.Join(constants.DefaultConfigPath, common.InstallerName+"."+common.InstallerConfigType)
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		// create config file
		file, err := os.Create(ConfigFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// write default config
		_, err = file.WriteString(_confSample)
		if err != nil {
			panic(err)
		}
	}

	config.InitSetup(*configFlag)
	config.SysRoot = sysRoot

	logger.LogInit(config.AppInfo.LogPath, config.AppInfo.LogSaveName, config.AppInfo.LogFileExt)

	service.MyService = service.NewService(config.CommonInfo.RuntimePath)
	go func() {
		var releaseURL []string
		for _, mirror := range config.ServerInfo.Mirrors {
			releaseURL = append(releaseURL, service.HyperFileTagReleaseURL(service.GetReleaseBranch(sysRoot), mirror))
		}
		var best service.BestURLFunc = service.BestByDelay // dependency inject
		best(releaseURL)
	}()
}

func main() {
	service.InstallerService = service.NewStatusService(service.NewInstallerService(sysRoot), sysRoot)

	err := service.InstallerService.Launch(sysRoot)
	if err != nil {
		logger.Error("error when trying to launch", zap.Error(err))
	}

	// watch rauc offline and release
	os.MkdirAll(config.RAUC_OFFLINE_PATH, os.ModePerm)
	os.MkdirAll(config.RAUC_RELEASE_PATH, os.ModePerm)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("offline create watcher error ", zap.Any("error", err))
	} else {
		defer watcher.Close()
	}
	watchOfflineDir(watcher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	MigrateOrFixOldVersion()

	mux := &util_http.HandlerMultiplexer{
		HandlerMap: map[string]http.Handler{
			"v2":  route.InitV2Router(),
			"doc": route.InitV2DocRouter(_docHTML, _docYAML),
		},
	}

	// notify systemd that we are ready
	{
		if supported, err := daemon.SdNotify(false, daemon.SdNotifyReady); err != nil {
			logger.Error("Failed to notify systemd that installer service is ready", zap.Any("error", err))
		} else if supported {
			logger.Info("Notified systemd that installer service is ready")
		} else {
			logger.Info("This process is not running as a systemd service.")
		}
	}

	// initialize routers and register at gateway
	listener, err := net.Listen("tcp", net.JoinHostPort(common.Localhost, "0"))
	if err != nil {
		panic(err)
	}

	go registerRouter(listener)

	// should do before cron job to prevent stop by `installing` status
	err = service.InstallerService.PostMigration(sysRoot)
	if err != nil {
		logger.Error("error when trying to post migration", zap.Error(err))
	}

	{
		// TODO 考虑重构程序的架构
		// 在最早，程序是 Event-Drive 的(不是我写的)。所有的数据请求都是在前端请求之后进行立刻获取的
		// 因为那时候的业务是一个安装器没有性能需求。
		// 但是ET说要改成一个OTA工具，所以我来接受这个项目
		// 所以导致了及时获取有一个性能上的延时，所以加入 Cron 来把数据提前缓存好。
		// 但是也有一个问题，缓存没有好的时候就访问 OTA 相关的接口会导致数据丢失。
		crontab := cron.New(cron.WithSeconds())

		go cronjob(ctx) // run once immediately

		if _, err := crontab.AddFunc("@every 30m", func() { cronjob(ctx) }); err != nil {
			logger.Error("error when trying to add cron job", zap.Error(err))
		}

		crontab.Start()
		defer crontab.Stop()
	}

	s := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // fix G112: Potential slowloris attack (see https://github.com/securego/gosec)
	}

	logger.Info("installer service is listening...", zap.String("address", listener.Addr().String()))
	if err := s.Serve(listener); err != nil {
		logger.Error("error when trying to serve", zap.Error(err))
		panic(err)
	}
}

func cronjob(ctx context.Context) {
	err := service.InstallerService.Cronjob(ctx, sysRoot)
	if err != nil {
		logger.Error("error when trying to cronjob", zap.Error(err))
	}
}

func watchOfflineDir(watcher *fsnotify.Watcher) {
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Create) {
					service.InstallerService = service.NewStatusService(service.NewInstallerService(sysRoot), sysRoot)
					service.InstallerService.Cronjob(context.Background(), sysRoot)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("offline watch err", zap.Any("error info", err))
			}
		}
	}()

	err := watcher.Add(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH))
	if err != nil {
		logger.Error("offline watch err", zap.Any("error info", err))
	}
}

func registerRouter(listener net.Listener) {
	for i := 0; i < 10; i++ {
		// initialize routers and register at gateway
		if gateway, err := service.MyService.Gateway(); err != nil {
			logger.Info("error when trying to connect to gateway... skipping", zap.Error(err))
		} else {
			apiPaths := []string{
				route.V2APIPath,
				route.V2DocPath,
			}

			for _, apiPath := range apiPaths {
				if err := gateway.CreateRoute(&model.Route{
					Path:   apiPath,
					Target: "http://" + listener.Addr().String(),
				}); err != nil {
					panic(err)
				}
			}
			logger.Info("gateway register success")
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func MigrateOrFixOldVersion() {
	Fix130WrongLatestDir()
}

func Fix130WrongLatestDir() {
	// bug ref: https://icewhale.feishu.cn/wiki/J9KBwpgmFiqjt8kv17DcAz1gnch

	if _, err := os.Stat(filepath.Join(config.SysRoot, filepath.Dir(config.RAUC_RELEASE_PATH), "latest")); err == nil {
		err := os.Remove(filepath.Join(config.SysRoot, filepath.Dir(config.RAUC_RELEASE_PATH), "latest"))
		if err != nil {
			logger.Error("error when trying to remove latest dir", zap.Error(err))
		}
	}
}
