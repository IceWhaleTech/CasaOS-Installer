package internal

import (
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"go.uber.org/zap"
)

func IsEmptyDir(path string) (bool, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	return len(files) == 0, nil
}

func CleanWithWhiteList(dir string, whiteList []string) error {
	for _, file := range whiteList {
		filePath := filepath.Join(dir, file)
		if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
		}
		if err := os.Remove(filePath); err != nil {
			logger.Error("error when trying to remove file", zap.Error(err))
			return err
		}
	}

	if isEmpty, err := IsEmptyDir(dir); err == nil && isEmpty {
		if err = os.Remove(dir); err != nil {
			logger.Error("error when trying to remove dir", zap.Error(err))
			return err
		}
	}

	return nil
}
