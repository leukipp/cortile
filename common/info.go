package common

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"net/http"
	"path/filepath"
)

var (
	Process ProcessInfo // Process information
	Build   BuildInfo   // Build information
)

type ProcessInfo struct {
	Id     int    // Process id
	Path   string // Process path
	Host   string // Process host
	System string // Process system
}

type BuildInfo struct {
	Name    string // Build name
	Version string // Build version
	Commit  string // Build commit
	Date    string // Build date
	Source  string // Build source
	Latest  string // Build latest
	Summary string // Build summary
}

func InitInfo(name, version, commit, date, source string) {

	// Process information
	Process = ProcessInfo{
		Id:     os.Getpid(),
		System: fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH),
	}
	Process.Host, _ = os.Hostname()
	Process.Path, _ = os.Executable()

	// Build information
	Build = BuildInfo{
		Name:    name,
		Version: version,
		Commit:  TruncateString(commit, 7),
		Date:    date,
		Source:  source,
		Latest:  version,
	}
	Build.Summary = fmt.Sprintf("%s v%s-%s, built on %s", Build.Name, Build.Version, Build.Commit, Build.Date)

	// Check latest version
	if !Feature("disable-version-check") && VersionToInt(Build.Version) > 0 {
		Build.Latest = Latest(source)
		if VersionToInt(Build.Latest) > VersionToInt(Build.Version) {
			Build.Summary = fmt.Sprintf("%s, >>> %s v%s available <<<", Build.Summary, Build.Name, Build.Latest)
		}
	}
}

func Latest(source string) string {

	// Request latest version from github
	res, err := http.Get(source + "/releases/latest")
	if err != nil {
		return Build.Version
	}

	// Parse latest version from redirect url
	version := path.Base(res.Request.URL.Path)
	if !strings.HasPrefix(version, "v") {
		return Build.Version
	}

	return version[1:]
}

func Feature(name string) bool {
	file := filepath.Join(Args.Cache, name)

	// Check if feature file exists
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		return true
	}
	return false
}
