package fixtures

import (
	"time"

	"github.com/IceWhaleTech/CasaOS-Installer/service"
)

func WaitFetchReleaseCompeleted(s *service.StatusService) {
	time.Sleep(service.GetReleaseCostTime + 500*time.Millisecond)
}

func WaitDownloadCompeleted(s *service.StatusService) {
	time.Sleep(service.DownloadCostTime + 500*time.Millisecond)
}
