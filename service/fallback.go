package service

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Installer/common"
)

// embeded static files
//
//go:embed static
var FallbackStaticFiles embed.FS
var srv *http.Server

func StartFallbackWebsite() {
	// run a http server for embedded fellbackStaticFiles
	fs, err := fs.Sub(FallbackStaticFiles, "static")
	if err != nil {
		panic(err)
	}

	port := "80"
	m := http.NewServeMux()
	srv = &http.Server{
		Addr:    net.JoinHostPort(common.Localhost, port),
		Handler: m,
	}
	m.Handle("/", http.FileServer(http.FS(fs)))

	// get file content string
	// if file not exist, use default config
	// to check exsit /etc/casaos/gateway.ini
	if content, err := os.ReadFile("/etc/casaos/gateway.ini"); err != nil {
		fmt.Println("read gateway config file error, use default config")
	} else {
		// get port=xx from content
		// if not exist, use default config
		strings.Split(string(content), "\n")
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "port=") {
				// run server in background
				port = strings.Split(line, "=")[1]
			}
		}

	}

	// to check port is a number
	if _, err := strconv.ParseInt(port, 10, 64); err != nil {
		fmt.Println("port is not a number, use default config")
		port = "80"
	}

	srv.Addr = net.JoinHostPort(common.Localhost, port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("start fallback website error: ", err)
	}

}

func StopFallbackWebsite() {
	count := 0
	for {
		if srv != nil {
			srv.Close()
			fmt.Println("stop fallback website")
			break
		} else {
			count++
		}
		if count > 5 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
