// Package upgrade provides version checking and upgrade functionality.
package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// GitHubAPIURL is the base URL for GitHub API
	GitHubAPIURL = "https://api.github.com"
	// DefaultOwner is the default repository owner
	DefaultOwner = "fregataa"
	// DefaultRepo is the default repository name
	DefaultRepo = "aami"
)

// Checker handles version checking against GitHub releases.
type Checker struct {
	owner      string
	repo       string
	httpClient *http.Client
}

// NewChecker creates a new version checker.
func NewChecker() *Checker {
	return &Checker{
		owner: DefaultOwner,
		repo:  DefaultRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithRepo sets a custom repository.
func (c *Checker) WithRepo(owner, repo string) *Checker {
	c.owner = owner
	c.repo = repo
	return c
}

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// VersionInfo contains parsed version information.
type VersionInfo struct {
	Version     string
	Major       int
	Minor       int
	Patch       int
	Prerelease  string
	IsValid     bool
}

// CheckResult contains the result of a version check.
type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	LatestRelease   *Release
	UpdateAvailable bool
	ReleaseNotes    string
	DownloadURL     string
}

// GetLatestRelease fetches the latest release from GitHub.
func (c *Checker) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GitHubAPIURL, c.owner, c.repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "aami-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &release, nil
}

// GetReleases fetches recent releases from GitHub.
func (c *Checker) GetReleases(limit int) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d", GitHubAPIURL, c.owner, c.repo, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "aami-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return releases, nil
}

// CheckForUpdate checks if an update is available.
func (c *Checker) CheckForUpdate(currentVersion string) (*CheckResult, error) {
	latest, err := c.GetLatestRelease()
	if err != nil {
		return nil, err
	}

	result := &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latest.TagName,
		LatestRelease:  latest,
		ReleaseNotes:   latest.Body,
	}

	// Find appropriate download URL
	result.DownloadURL = c.findDownloadURL(latest)

	// Compare versions
	current := ParseVersion(currentVersion)
	latestVer := ParseVersion(latest.TagName)

	result.UpdateAvailable = latestVer.IsNewer(current)

	return result, nil
}

// findDownloadURL finds the appropriate binary for the current platform.
func (c *Checker) findDownloadURL(release *Release) string {
	// Look for platform-specific binary
	// Priority: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
	patterns := []string{
		"linux-amd64",
		"linux-arm64",
		"darwin-amd64",
		"darwin-arm64",
	}

	for _, pattern := range patterns {
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, pattern) && !strings.HasSuffix(asset.Name, ".sha256") {
				return asset.BrowserDownloadURL
			}
		}
	}

	// Return first non-checksum asset as fallback
	for _, asset := range release.Assets {
		if !strings.HasSuffix(asset.Name, ".sha256") && !strings.HasSuffix(asset.Name, ".md5") {
			return asset.BrowserDownloadURL
		}
	}

	return release.HTMLURL
}

// ParseVersion parses a version string into VersionInfo.
func ParseVersion(version string) VersionInfo {
	info := VersionInfo{Version: version}

	// Remove 'v' prefix if present
	v := strings.TrimPrefix(version, "v")

	// Match semver pattern: major.minor.patch[-prerelease]
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)
	matches := re.FindStringSubmatch(v)

	if matches == nil {
		// Try simple version without patch
		re = regexp.MustCompile(`^(\d+)\.(\d+)(?:-(.+))?$`)
		matches = re.FindStringSubmatch(v)
		if matches != nil {
			info.Major, _ = strconv.Atoi(matches[1])
			info.Minor, _ = strconv.Atoi(matches[2])
			if len(matches) > 3 {
				info.Prerelease = matches[3]
			}
			info.IsValid = true
		}
		return info
	}

	info.Major, _ = strconv.Atoi(matches[1])
	info.Minor, _ = strconv.Atoi(matches[2])
	info.Patch, _ = strconv.Atoi(matches[3])
	if len(matches) > 4 {
		info.Prerelease = matches[4]
	}
	info.IsValid = true

	return info
}

// IsNewer returns true if v is newer than other.
func (v VersionInfo) IsNewer(other VersionInfo) bool {
	if !v.IsValid || !other.IsValid {
		return false
	}

	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}

	// If versions are equal, non-prerelease is newer than prerelease
	if v.Prerelease == "" && other.Prerelease != "" {
		return true
	}

	return false
}

// String returns the version string.
func (v VersionInfo) String() string {
	if v.Prerelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
