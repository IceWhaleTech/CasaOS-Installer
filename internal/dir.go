package internal

import (
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"go.uber.org/zap"
)

func IsDirExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func IsEmptyDir(path string) (bool, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	return len(files) == 0, nil
}

func CleanWithWhiteList(dir string, whiteList []string, cleanDir bool) error {
	for _, file := range whiteList {
		if err := os.Remove(filepath.Join(dir, file)); err != nil {
			logger.Error("error when trying to remove file", zap.Error(err))
		}
	}

	isEmpty, err := IsEmptyDir(dir)
	if err != nil {
		logger.Error("error when trying to check if dir is empty", zap.Error(err))
	}

	if isEmpty && cleanDir {
		if err = os.Remove(dir); err != nil {
			logger.Error("error when trying to remove dir", zap.Error(err))
		}

		if IsDirExist(dir) {
			logger.Error("error when trying to remove dir", zap.Error(err))
		}
	}

	return nil
}
