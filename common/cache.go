package common

import (
	"os"
	"strings"

	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Cache[T any] struct {
	Folder string // Cache file folder
	Name   string // Cache file name
	Data   T      // Cache file data
}

func InitCache() {
	if HasFlag("disable-cache-folder") {
		Args.Cache = "disabled"
	}
	if CacheDisabled() {
		return
	}

	// Create cache folder if not exists
	cacheFolderPath := Args.Cache
	if _, err := os.Stat(cacheFolderPath); os.IsNotExist(err) {
		os.MkdirAll(cacheFolderPath, 0755)
	}
}

func CacheFolderPath(name string) string {

	// Obtain user cache directory
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Error obtaining cache directory: ", err)
	}

	return filepath.Join(userCacheDir, name)
}

func CacheDisabled() bool {
	arg := strings.ToLower(strings.TrimSpace(Args.Cache))
	return IsInList(arg, []string{"", "0", "off", "false", "disabled"})
}
