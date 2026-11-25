package context

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GitInfo contains repository information
type GitInfo struct {
	RepoURL    string
	CommitHash string
}

// errGitUnavailable is returned when neither the remote URL nor the commit could be read
var errGitUnavailable = errors.New("git repository information unavailable")

// scpLikeRegex matches scp-like SSH remotes such as git@github.com:org/repo.git
var scpLikeRegex = regexp.MustCompile(`^([A-Za-z0-9._-]+@)?([A-Za-z0-9.-]+):([^/].*)$`)

var (
	gitCache     *GitInfo
	gitCacheErr  error
	gitCacheLock sync.RWMutex
	gitCacheTime time.Time
	gitCacheTTL  = 5 * time.Minute
)

// gitCommand runs git with the given arguments and returns its trimmed output.
// It is a variable so tests can substitute a fake implementation.
var gitCommand = func(args ...string) (string, error) {
	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetGitInfo retrieves git repository information with caching.
// It returns errGitUnavailable when neither the remote URL nor the commit can be read.
// Both successful and failed lookups are cached for the TTL.
func GetGitInfo() (*GitInfo, error) {
	gitCacheLock.RLock()
	if !gitCacheTime.IsZero() && time.Since(gitCacheTime) < gitCacheTTL {
		info, err := copyCachedGitInfo()
		gitCacheLock.RUnlock()
		return info, err
	}
	gitCacheLock.RUnlock()

	// Need to fetch new info
	gitCacheLock.Lock()
	defer gitCacheLock.Unlock()

	// Check again in case another goroutine updated it
	if !gitCacheTime.IsZero() && time.Since(gitCacheTime) < gitCacheTTL {
		return copyCachedGitInfo()
	}

	info := &GitInfo{}

	// Get repository URL; credentials are always stripped
	if repoURL, err := gitCommand("config", "--get", "remote.origin.url"); err == nil {
		if normalized, err := normalizeRepoURL(repoURL); err == nil {
			info.RepoURL = normalized
		}
	}

	// Get commit hash
	if commit, err := gitCommand("rev-parse", "HEAD"); err == nil {
		info.CommitHash = commit
	}

	var err error
	if info.RepoURL == "" && info.CommitHash == "" {
		err = errGitUnavailable
	}

	// Update cache
	gitCache = info
	gitCacheErr = err
	gitCacheTime = time.Now()

	return copyCachedGitInfo()
}

// copyCachedGitInfo returns a copy of the cached value; callers must hold gitCacheLock
func copyCachedGitInfo() (*GitInfo, error) {
	if gitCacheErr != nil {
		return nil, gitCacheErr
	}
	info := *gitCache
	return &info, nil
}

// normalizeRepoURL converts a git remote URL to a public, credential-free form.
// SSH remotes are converted to HTTPS, any user info and port in SSH form are dropped,
// and a trailing .git suffix is removed.
func normalizeRepoURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty repository URL")
	}

	// Rewrite scp-like syntax (git@host:path) to an ssh:// URL so it can be parsed
	if !strings.Contains(raw, "://") {
		m := scpLikeRegex.FindStringSubmatch(raw)
		if m == nil {
			return "", fmt.Errorf("unrecognized repository URL: %s", raw)
		}
		raw = "ssh://" + m[1] + m[2] + "/" + m[3]
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %w", err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("repository URL has no host: %s", raw)
	}

	if u.Scheme == "ssh" || u.Scheme == "git+ssh" {
		u.Scheme = "https"
		u.Host = u.Hostname()
	}

	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimSuffix(u.Path, ".git")
	u.RawPath = ""

	return u.String(), nil
}

// ClearGitCache clears the git information cache
func ClearGitCache() {
	gitCacheLock.Lock()
	defer gitCacheLock.Unlock()
	gitCache = nil
	gitCacheErr = nil
	gitCacheTime = time.Time{}
}
