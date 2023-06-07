//go:generate bash -c "mkdir -p codegen && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,server,spec -package codegen api/openapi.yaml > codegen/api.go"

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/IceWhaleTech/CasaOS-Installer/common"
	"github.com/IceWhaleTech/CasaOS-Installer/internal/config"
	"github.com/coreos/go-systemd/daemon"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	commit = "private build"
	date   = "private build"
)

func main() {
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
	}

	// TODO: setup cron to check for new release periodically
	{
		crontab := cron.New(cron.WithSeconds())

		go func() {
			// TODO: run once at startup
		}()

		if _, err := crontab.AddFunc("@every 24h", func() {
			// TODO: run every 24 hours
		}); err != nil {
			panic(err)
		}

		crontab.Start()
		defer crontab.Stop()
	}

	apiService, apiServiceError := StartAPIService()

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

	// Set up a channel to catch the Ctrl+C signal (SIGINT)
	{
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		// Wait for the signal or server error
		select {
		case <-signalChan:
			fmt.Println("\nReceived signal, shutting down server...")
		case err := <-apiServiceError:
			fmt.Printf("Error starting API service: %s\n", err)
			if err != http.ErrServerClosed {
				os.Exit(1)
			}
		}
	}

	// Create a context with a timeout to allow the server to shut down gracefully
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown the apiService
		if err := apiService.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown api server", zap.Any("error", err))
			os.Exit(1)
		}
	}
}
