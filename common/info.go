package common

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"encoding/json"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/shirou/gopsutil/host"
)

var (
	Process ProcessInfo // Process information
	Build   BuildInfo   // Build information
	Source  SourceInfo  // Source information
)

type ProcessInfo struct {
	Id   int            // Process id
	Path string         // Process path
	User *user.User     // Process user
	Host *host.InfoStat // Process host
}

type BuildInfo struct {
	Name    string // Build name
	Target  string // Build target
	Version string // Build version
	Commit  string // Build commit
	Date    string // Build date
	Summary string // Build summary
}

type SourceInfo struct {
	Hostname   string // Source code hostname
	Repository string // Source code repository
	Releases   []Info // Source code releases
	Issues     []Info // Source code issues
}

type Info struct {
	Id      int    // Source item id
	Url     string // Source item url
	Name    string // Source item name
	Created string // Source item date
	Type    string // Source item type
	Extra   *Info  // Source extra info
}

func InitInfo(name, target, version, commit, date, source string) {

	// Process information
	Process = ProcessInfo{
		Id: os.Getpid(),
	}

	// Process path information
	processPath, err := os.Executable()
	if err != nil {
		processPath = "unknown"
	}
	Process.Path = processPath

	// Process user information
	processUser, err := user.Current()
	if err != nil {
		processUser = &user.User{Username: "unknown", Name: "unknown"}
	}
	Process.User = processUser

	// Process host information
	processHost, err := host.Info()
	if err != nil {
		processHost = &host.InfoStat{Hostname: "unknown", Platform: "unknown"}
	}
	Process.Host = processHost

	// Build information
	Build = BuildInfo{
		Name:    name,
		Target:  target,
		Version: version,
		Commit:  TruncateString(commit, 7),
		Date:    date,
	}
	Build.Summary = fmt.Sprintf("%s v%s-%s, built on %s (%s)", Build.Name, Build.Version, Build.Commit, Build.Date, Build.Target)

	// Source information
	hostname, repository, _ := strings.Cut(source, "/")
	Source = SourceInfo{
		Hostname:   hostname,
		Repository: repository,
		Releases:   FetchReleases(hostname, repository),
		Issues:     FetchIssues(hostname, repository, "info"),
	}

	// Update build summary
	if HasReleaseInfos() {
		Build.Summary = fmt.Sprintf("%s, >>> %s v%s is available <<<", Build.Summary, Build.Name, Source.Releases[0].Name)
	}
}

func FetchReleases(hostname, repository string) []Info {
	releases := []Info{}
	if HasFlag("disable-release-info") || IsDevVersion() {
		return releases
	}

	// Request latest release
	response, err := http.Get(fmt.Sprintf("https://api.%s/repos/%s/releases/latest", hostname, repository))
	if err != nil {
		return releases
	}

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return releases
	}

	// Parse response body
	data := Map{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return releases
	}

	// Parse release data
	if !IsInMap(data, []string{"id", "html_url", "tag_name", "created_at"}) {
		return releases
	}
	latest := Info{
		Id:      int(data["id"].(float64)),
		Url:     data["html_url"].(string),
		Name:    data["tag_name"].(string)[1:],
		Created: data["created_at"].(string),
		Type:    "releases",
	}

	// Add asset information
	assetName := fmt.Sprintf(
		"%s_%s_%s.tar.gz",
		Build.Name,
		latest.Name,
		strings.Replace(Build.Target, "-", "_", -1),
	)
	assetUrl := fmt.Sprintf(
		"https://%s/%s/releases/download/v%s/%s",
		hostname,
		repository,
		latest.Name,
		assetName,
	)
	asset := &Info{
		Id:      latest.Id,
		Url:     assetUrl,
		Name:    assetName,
		Created: latest.Created,
		Type:    "assets",
	}
	latest.Extra = asset

	// Add checksum information
	checksumName := fmt.Sprintf(
		"%s_%s_checksums.txt",
		Build.Name,
		latest.Name,
	)
	checksumUrl := fmt.Sprintf(
		"https://%s/%s/releases/download/v%s/%s",
		hostname,
		repository,
		latest.Name,
		checksumName,
	)
	checksum := &Info{
		Id:      asset.Id,
		Url:     checksumUrl,
		Name:    checksumName,
		Created: asset.Created,
		Type:    "checksums",
	}
	asset.Extra = checksum

	return append(releases, latest)

}

func FetchIssues(hostname, repository, labels string) []Info {
	issues := []Info{}
	if HasFlag("disable-issue-info") || IsDevVersion() {
		return issues
	}

	// Request repository issues
	response, err := http.Get(fmt.Sprintf("https://api.%s/repos/%s/issues?labels=%s", hostname, repository, labels))
	if err != nil {
		return issues
	}

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return issues
	}

	// Parse response body
	data := List{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return issues
	}

	// Parse issue data
	for _, item := range data {
		if !IsInMap(item, []string{"number", "html_url", "title", "created_at"}) {
			continue
		}
		issues = append(issues, Info{
			Id:      int(item["number"].(float64)),
			Url:     item["html_url"].(string),
			Name:    item["title"].(string),
			Created: item["created_at"].(string),
			Type:    "issues",
		})
	}

	return issues
}

func VersionToInt(version string) int {

	// Remove non-numeric characters
	reg := regexp.MustCompile("[^0-9]+")
	numeric := reg.ReplaceAllString(strings.Split(version, "-")[0], "")

	// Convert version string to integer
	integer, err := strconv.Atoi(numeric)
	if err != nil {
		return -1
	}

	return integer
}

func IsDevVersion() bool {
	return VersionToInt(Build.Version) < 1
}

func HasFlag(name string) bool {
	return IsInList(name, os.Args[1:])
}

func HasReleaseInfos() bool {
	return len(Source.Releases) > 0 && VersionToInt(Source.Releases[0].Name) > VersionToInt(Build.Version)
}

func HasIssueInfos() bool {
	return len(Source.Issues) > 0
}

func HasUnseenInfos() bool {
	unseen := false

	// Check for uncached release infos
	if HasReleaseInfos() {
		for _, release := range Source.Releases {
			unseen = unseen || release.Unseen()
		}
	}

	// Check for uncached issue infos
	if HasIssueInfos() {
		for _, issue := range Source.Issues {
			unseen = unseen || issue.Unseen()
		}
	}

	return unseen
}

func SemverUpdateInfos() (bool, bool, bool) {
	if !HasReleaseInfos() {
		return false, false, false
	}

	// Split version into major, minor and patch slice
	update := StringsToInts(strings.Split(Source.Releases[0].Name, "."))
	current := StringsToInts(strings.Split(Build.Version, "."))
	if len(current) != 3 || len(update) != 3 {
		return false, false, false
	}

	return update[0] > current[0], update[1] > current[1], update[2] > current[2]
}

func (i *Info) Unseen() bool {
	if CacheDisabled() {
		return false
	}

	// Check info cache file
	cache := i.Cache()
	_, err := os.Stat(filepath.Join(cache.Folder, cache.Name))

	return os.IsNotExist(err)
}

func (i *Info) Seen() bool {
	if CacheDisabled() {
		return false
	}

	// Obtain cache object
	cache := i.Cache()

	// Parse info cache
	data, err := json.MarshalIndent(cache.Data, "", "  ")
	if err != nil {
		return false
	}

	// Write info cache
	path := filepath.Join(cache.Folder, cache.Name)
	err = os.WriteFile(path, data, 0644)

	return err == nil
}

func (i *Info) Cache() Cache[*Info] {
	subfolder := strings.ToLower(i.Type)
	filename := fmt.Sprintf("%s-%s-%d", subfolder, i.Name, i.Id)

	// Create info cache folder
	folder := filepath.Join(Args.Cache, "infos", subfolder)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0755)
	}

	// Create info cache object
	cache := Cache[*Info]{
		Folder: folder,
		Name:   HashString(filename, 20) + ".json",
		Data:   i,
	}

	return cache
}
