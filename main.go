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
	"sync"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"

	util_http "github.com/IceWhaleTech/CasaOS-Common/utils/http"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen"
	"github.com/IceWhaleTech/CasaOS-Installer/codegen/message_bus"
	"github.com/IceWhaleTech/CasaOS-Installer/internal"

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
			releaseURL = append(releaseURL, service.HyperFileTagReleaseUrl(service.GetReleaseBranch(sysRoot), mirror))
		}
		var best service.BestURLFunc = service.BestByDelay // dependency inject
		best(releaseURL)
	}()
}

func main() {

	service.InstallerService = &service.StatusService{
		ImplementService:                 service.NewInstallerService(sysRoot),
		SysRoot:                          sysRoot,
		Have_other_get_release_flag_lock: sync.RWMutex{},
	}

	service.InstallerService.Launch(sysRoot)

	// watch rauc offline
	os.MkdirAll(config.RAUC_OFFLINE_PATH, os.ModePerm)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("offline create watcher error ", zap.Any("error", err))
	} else {
		defer watcher.Close()
	}
	watchOfflineDir(watcher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	go registerMsg()

	{
		crontab := cron.New(cron.WithSeconds())

		go cronjob(ctx) // run once immediately

		if _, err := crontab.AddFunc("@every 60m", func() { cronjob(ctx) }); err != nil {
			logger.Error("error when trying to add cron job", zap.Error(err))
		}

		crontab.Start()
		defer crontab.Stop()
	}

	service.InstallerService.PostMigration(sysRoot)

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
	go internal.DownloadReleaseBackground(*release.Background, release.Version)

	if err != nil {
		logger.Error("error when trying to get release", zap.Error(err))
		return
	}
	// cache release packages if not already cached
	shouldUpgrade := service.InstallerService.ShouldUpgrade(*release, sysRoot)
	if shouldUpgrade {
		releaseFilePath, err := service.InstallerService.DownloadRelease(ctx, *release, true)
		if err != nil {
			logger.Error("error when trying to download release", zap.Error(err), zap.String("release file path", releaseFilePath))
			return
		}
	}
	//isUpgradable := service.InstallerService.IsUpgradable(*release, sysRoot)

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
					service.InstallerService = &service.StatusService{
						ImplementService:                 service.NewInstallerService(sysRoot),
						SysRoot:                          sysRoot,
						Have_other_get_release_flag_lock: sync.RWMutex{},
					}
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
func registerMsg() {
	var messageBus *message_bus.ClientWithResponses
	var err error
	for i := 0; i < 10; i++ {
		if messageBus, err = service.MyService.MessageBus(); err != nil {
			logger.Error("error when trying to connect to message bus... skipping", zap.Error(err))
			continue
		}
		response, err := messageBus.RegisterEventTypesWithResponse(context.Background(), common.EventTypes)
		if err != nil {
			logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.Error(err))
			continue
		}
		if response != nil && response.StatusCode() != http.StatusOK {
			logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.String("status", response.Status()), zap.String("body", string(response.Body)))
			continue
		}
		time.Sleep(3 * time.Second)
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
