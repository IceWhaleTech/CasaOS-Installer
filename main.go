//go:generate bash -c "mkdir -p codegen && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,server,spec -package codegen api/openapi.yaml > codegen/api.go"
//go:generate bash -c "mkdir -p codegen/message_bus && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,client -package message_bus https://raw.githubusercontent.com/IceWhaleTech/CasaOS-MessageBus/main/api/message_bus/openapi.yaml > codegen/message_bus/api.go"

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
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/IceWhaleTech/CasaOS-Installer/route"
	"github.com/IceWhaleTech/CasaOS-Installer/service"
	"github.com/IceWhaleTech/CasaOS-Installer/types"
	"github.com/coreos/go-systemd/daemon"
	"go.uber.org/zap"
)

var (
	commit = "private build"
	date   = "private build"

	//go:embed api/index.html
	_docHTML string

	//go:embed api/openapi.yaml
	_docYAML string
	sysRoot  = "/"

	//go:embed build/sysroot/etc/casaos/installer.conf.sample
	_confSample string
)

func main() {
	// create config
	{
		// create default config file if not exist
		ConfigFilePath := filepath.Join(constants.DefaultConfigPath, common.InstallerName+"."+common.InstallerConfigType)
		if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
			println("config file not exist, create it")
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
	}

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
		config.SysRoot = sysRoot

		logger.LogInit(config.AppInfo.LogPath, config.AppInfo.LogSaveName, config.AppInfo.LogFileExt)

		service.MyService = service.NewService(config.CommonInfo.RuntimePath)
	}

	service.InstallerService = &service.StatusService{
		ImplementService: service.NewInstallerService(sysRoot),
		SysRoot:          sysRoot,
	}

	go service.StartFallbackWebsite()

	service.InstallerService.Launch(sysRoot)

	// watch rauc offline
	{
		// 这个是临时放这里，为了watch里不会没有东西。
		err := os.MkdirAll(config.RAUC_OFFLINE_PATH, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			os.MkdirAll(config.RAUC_OFFLINE_PATH, os.ModePerm)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logger.Error(err.Error())
			panic(err)
		}
		defer watcher.Close()

		// Start listening for events.
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if event.Has(fsnotify.Create) {
						service.InstallerService = &service.StatusService{
							ImplementService: service.NewInstallerService(sysRoot),
							SysRoot:          sysRoot,
						}
					}
					if event.Has(fsnotify.Remove) {
						service.InstallerService = &service.StatusService{
							ImplementService: service.NewInstallerService(sysRoot),
							SysRoot:          sysRoot,
						}
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					logger.Error(err.Error())
				}
			}
		}()

		// Add a path.
		err = watcher.Add(filepath.Join(sysRoot, config.RAUC_OFFLINE_PATH))
		if err != nil {
			logger.Error(err.Error())
			os.MkdirAll(config.RAUC_OFFLINE_PATH, os.ModePerm)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := &util_http.HandlerMultiplexer{
		HandlerMap: map[string]http.Handler{
			"v2":  route.InitV2Router(),
			"doc": route.InitV2DocRouter(_docHTML, _docYAML),
		},
	}

	service.StopFallbackWebsite()

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

	// 上面notify之后，才有必要去注册
	go func() {
		for {
			// register at message bus
			if messageBus, err := service.MyService.MessageBus(); err != nil {
				logger.Info("error when trying to connect to message bus... skipping", zap.Error(err))
			} else {
				response, err := messageBus.RegisterEventTypesWithResponse(ctx, common.EventTypes)
				if err != nil {
					logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.Error(err))
				}

				if response != nil && response.StatusCode() != http.StatusOK {
					logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.String("status", response.Status()), zap.String("body", string(response.Body)))
				}
				break
			}
		}
		// 不要一直重试
		time.Sleep(10 * time.Second)
	}()

	// initialize routers and register at gateway
	listener, err := net.Listen("tcp", net.JoinHostPort(common.Localhost, "0"))
	if err != nil {
		panic(err)
	}

	go func() {
		for {
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
			time.Sleep(10 * time.Second)
		}
	}()

	// 等待一下，让gateway注册成功
	time.Sleep(5 * time.Second)
	// 这里应该还要把文件删一下
	service.InstallerService.PostMigration(sysRoot)

	s := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // fix G112: Potential slowloris attack (see https://github.com/securego/gosec)
	}

	{
		crontab := cron.New(cron.WithSeconds())

		go cronjob(ctx) // run once immediately

		if _, err := crontab.AddFunc("@every 20m", func() { cronjob(ctx) }); err != nil {
			panic(err)
		}

		// every 10 seconds for debug
		if _, err := crontab.AddFunc("@every 1s", func() {
		}); err != nil {
			panic(err)
		}

		crontab.Start()
		defer crontab.Stop()
	}

	logger.Info("installer service is listening...", zap.String("address", listener.Addr().String()))
	if err := s.Serve(listener); err != nil {
		panic(err)
	}

}

func cronjob(ctx context.Context) {

	status, _ := service.GetStatus()
	if status.Status == codegen.Downloading {
		return
	}

	if status.Status == codegen.Installing {
		return
	}

	// release, err := service.GetRelease(ctx, service.GetReleaseBranch(sysRoot))
	ctx = context.WithValue(ctx, types.Trigger, types.CRON_JOB)
	release, err := service.InstallerService.GetRelease(ctx, service.GetReleaseBranch(sysRoot))

	if err != nil {
		logger.Error("error when trying to get release", zap.Error(err))
		return
	}

	// cache release packages if not already cached
	shouldUpgrade := service.InstallerService.ShouldUpgrade(*release, sysRoot)
	isUpgradable := service.InstallerService.IsUpgradable(*release, sysRoot)
	if shouldUpgrade && !isUpgradable {

		logger.Info("error while verifying release - continue to download", zap.Error(err))

		releaseFilePath, err := service.InstallerService.DownloadRelease(ctx, *release, true)

		if err != nil {
			logger.Error("error when trying to download release", zap.Error(err))
			return
		}
		logger.Info("downloaded release", zap.String("release file path", releaseFilePath))
	} else {
		// fmt.Println("不需要更新")
		// fmt.Println("service.ShouldUpgrade", shouldUpgrade)
		// fmt.Println("service.IsUpgradable", isUpgradable)
	}
}
