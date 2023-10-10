package common

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"net/http"
	"path/filepath"
)

var (
	Build BuildInfo // Build information
	Args  Arguments // Parsed arguments
)

type BuildInfo struct {
	Name    string // Build name
	Version string // Build version
	Commit  string // Build commit
	Date    string // Build date
	Source  string // Build summary
	Latest  string // Build latest
	Summary string // Build summary
}

type Arguments struct {
	Cache  string // Argument for cache folder path
	Config string // Argument for config file path
	Lock   string // Argument for lock file path
	Sock   string // Argument for sock file path
	Log    string // Argument for log file path
	VVV    bool   // Argument for very very verbose mode
	VV     bool   // Argument for very verbose mode
	V      bool   // Argument for verbose mode
}

func InitArgs(name, version, commit, date, source string) {

	// Build information
	Build = BuildInfo{
		Name:    name,
		Version: version,
		Commit:  Truncate(commit, 7),
		Date:    date,
		Source:  source,
		Latest:  version,
	}
	Build.Summary = fmt.Sprintf("%s v%s-%s, built on %s", Build.Name, Build.Version, Build.Commit, Build.Date)

	// Command line arguments
	flag.StringVar(&Args.Cache, "cache", filepath.Join(CacheFolderPath(Build.Name), Build.Version), "cache folder path")
	flag.StringVar(&Args.Config, "config", filepath.Join(ConfigFolderPath(Build.Name), "config.toml"), "config file path")
	flag.StringVar(&Args.Lock, "lock", filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", Build.Name)), "lock file path")
	flag.StringVar(&Args.Sock, "sock", filepath.Join(os.TempDir(), fmt.Sprintf("%s.sock", Build.Name)), "sock file path")
	flag.StringVar(&Args.Log, "log", filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", Build.Name)), "log file path")
	flag.BoolVar(&Args.VVV, "vvv", false, "very very verbose mode")
	flag.BoolVar(&Args.VV, "vv", false, "very verbose mode")
	flag.BoolVar(&Args.V, "v", false, "verbose mode")

	// Command line usage text
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", Build.Summary)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Version checker
	suspended := false
	if _, err := os.Stat(filepath.Join(Args.Cache, "no-version-check")); !os.IsNotExist(err) {
		suspended = true
	}
	if !suspended && VersionToInt(Build.Version) > 0 {
		Build.Latest = Latest(source)
		if VersionToInt(Build.Latest) > VersionToInt(Build.Version) {
			Build.Summary = fmt.Sprintf("%s, >>> %s v%s available <<<", Build.Summary, Build.Name, Build.Latest)
		}
	}
}

func Latest(source string) string {

	// Request latest version from github
	res, err := http.Get(strings.Trim(source, "/") + "/releases/latest")
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
