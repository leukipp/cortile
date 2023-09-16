package common

import (
	"os"

	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Cache[T any] struct {
	Folder string // Cache file folder
	Name   string // Cache file name
	Data   T      // Cache file data
}

func InitCache() {

	// Create cache folder if not exists
	cacheFolderPath := Args.Cache
	if _, err := os.Stat(cacheFolderPath); os.IsNotExist(err) {
		os.MkdirAll(cacheFolderPath, 0700)
	}
}

func CacheFolderPath(name string) string {

	// Obtain user cache directory
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Error obtaining cache directory ", err)
	}

	return filepath.Join(userCacheDir, name)
}
