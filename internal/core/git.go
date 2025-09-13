package core

import (
	"os/exec"
	"strings"
	"sync"
	"time"
)

// GitInfo contains repository information
type GitInfo struct {
	RepoURL    string
	CommitHash string
}

var (
	gitCache     *GitInfo
	gitCacheLock sync.RWMutex
	gitCacheTime time.Time
	gitCacheTTL  = 5 * time.Minute
)

// GetGitInfo retrieves git repository information with caching
func GetGitInfo() (*GitInfo, error) {
	gitCacheLock.RLock()
	if gitCache != nil && time.Since(gitCacheTime) < gitCacheTTL {
		info := *gitCache
		gitCacheLock.RUnlock()
		return &info, nil
	}
	gitCacheLock.RUnlock()

	// Need to fetch new info
	gitCacheLock.Lock()
	defer gitCacheLock.Unlock()

	// Check again in case another goroutine updated it
	if gitCache != nil && time.Since(gitCacheTime) < gitCacheTTL {
		info := *gitCache
		return &info, nil
	}

	info := &GitInfo{}

	// Get repository URL
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err == nil {
		repoURL := strings.TrimSpace(string(output))
		info.RepoURL = convertSSHToHTTPS(repoURL)
	}

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	output, err = cmd.Output()
	if err == nil {
		info.CommitHash = strings.TrimSpace(string(output))
	}

	// Update cache
	gitCache = info
	gitCacheTime = time.Now()

	return info, nil
}

// convertSSHToHTTPS converts SSH git URLs to HTTPS format
func convertSSHToHTTPS(url string) string {
	// Handle git@github.com:user/repo.git format
	if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
		url = "https://" + url
		url = strings.TrimSuffix(url, ".git")
		return url
	}

	// Handle ssh://git@bitbucket.org/user/repo.git format
	if strings.HasPrefix(url, "ssh://") {
		url = strings.TrimPrefix(url, "ssh://")
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
		url = "https://" + url
		url = strings.TrimSuffix(url, ".git")
		return url
	}

	// Already HTTPS or other format
	url = strings.TrimSuffix(url, ".git")
	return url
}

// ClearGitCache clears the git information cache
func ClearGitCache() {
	gitCacheLock.Lock()
	defer gitCacheLock.Unlock()
	gitCache = nil
	gitCacheTime = time.Time{}
}
