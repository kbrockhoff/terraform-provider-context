package context

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNormalizeRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "github scp-like ssh",
			input: "git@github.com:org/repo.git",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "gitlab scp-like ssh",
			input: "git@gitlab.com:user/repo.git",
			want:  "https://gitlab.com/user/repo",
		},
		{
			name:  "bitbucket ssh url",
			input: "ssh://git@bitbucket.org/org/repo.git",
			want:  "https://bitbucket.org/org/repo",
		},
		{
			name:  "ssh url with port",
			input: "ssh://git@host:2222/org/repo.git",
			want:  "https://host/org/repo",
		},
		{
			name:  "https with user and token",
			input: "https://user:token@github.com/org/repo.git",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "https with user only",
			input: "https://user@github.com/org/repo",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "https with git suffix",
			input: "https://github.com/org/repo.git",
			want:  "https://github.com/org/repo",
		},
		{
			name:  "plain http",
			input: "http://host/org/repo",
			want:  "http://host/org/repo",
		},
		{
			name:  "surrounding whitespace",
			input: "  https://github.com/org/repo.git\n",
			want:  "https://github.com/org/repo",
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "not a url",
			input:   "::not a url",
			wantErr: true,
		},
		{
			name:    "local path",
			input:   "/srv/git/repo.git",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeRepoURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeRepoURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizeRepoURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// stubGitCommand replaces gitCommand for the duration of a test
func stubGitCommand(t *testing.T, fn func(args ...string) (string, error)) {
	t.Helper()
	original := gitCommand
	gitCommand = fn
	ClearGitCache()
	t.Cleanup(func() {
		gitCommand = original
		ClearGitCache()
	})
}

func TestGetGitInfo(t *testing.T) {
	t.Run("strips credentials from remote", func(t *testing.T) {
		stubGitCommand(t, func(args ...string) (string, error) {
			if args[0] == "config" {
				return "https://user:secret@github.com/org/repo.git", nil
			}
			return "abc123", nil
		})

		info, err := GetGitInfo()
		if err != nil {
			t.Fatalf("GetGitInfo() error = %v", err)
		}
		if info.RepoURL != "https://github.com/org/repo" {
			t.Errorf("RepoURL = %q, want credential-free URL", info.RepoURL)
		}
		if info.CommitHash != "abc123" {
			t.Errorf("CommitHash = %q, want abc123", info.CommitHash)
		}
	})

	t.Run("unavailable outside a repository", func(t *testing.T) {
		stubGitCommand(t, func(args ...string) (string, error) {
			return "", errors.New("not a git repository")
		})

		info, err := GetGitInfo()
		if !errors.Is(err, errGitUnavailable) {
			t.Fatalf("GetGitInfo() error = %v, want errGitUnavailable", err)
		}
		if info != nil {
			t.Errorf("GetGitInfo() info = %+v, want nil", info)
		}
	})

	t.Run("commit only when remote is missing", func(t *testing.T) {
		stubGitCommand(t, func(args ...string) (string, error) {
			if args[0] == "config" {
				return "", errors.New("no remote")
			}
			return "def456", nil
		})

		info, err := GetGitInfo()
		if err != nil {
			t.Fatalf("GetGitInfo() error = %v", err)
		}
		if info.RepoURL != "" || info.CommitHash != "def456" {
			t.Errorf("GetGitInfo() = %+v, want commit only", info)
		}
	})

	t.Run("returns copies of cached value", func(t *testing.T) {
		stubGitCommand(t, func(args ...string) (string, error) {
			return "https://github.com/org/repo", nil
		})

		first, _ := GetGitInfo()
		first.RepoURL = "mutated"
		second, _ := GetGitInfo()
		if second.RepoURL == "mutated" {
			t.Error("cached GitInfo was mutated through a returned pointer")
		}
	})
}

func TestGetGitInfoConcurrentSingleLookup(t *testing.T) {
	var configCalls, revParseCalls atomic.Int32
	stubGitCommand(t, func(args ...string) (string, error) {
		switch args[0] {
		case "config":
			configCalls.Add(1)
			return "git@github.com:org/repo.git", nil
		case "rev-parse":
			revParseCalls.Add(1)
			return "abc123", nil
		}
		return "", errors.New("unexpected git command")
	})

	const workers = 100
	results := make([]*GitInfo, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			info, err := GetGitInfo()
			if err != nil {
				t.Errorf("GetGitInfo() error = %v", err)
				return
			}
			results[i] = info
		}(i)
	}
	wg.Wait()

	if got := configCalls.Load(); got != 1 {
		t.Errorf("git config called %d times, want 1", got)
	}
	if got := revParseCalls.Load(); got != 1 {
		t.Errorf("git rev-parse called %d times, want 1", got)
	}
	for i, info := range results {
		if info == nil || info.RepoURL != "https://github.com/org/repo" || info.CommitHash != "abc123" {
			t.Errorf("result %d = %+v, want identical cached values", i, info)
		}
	}
}

func TestClearGitCache(t *testing.T) {
	// Set up cache
	gitCache = &GitInfo{
		RepoURL:    "https://github.com/test/repo",
		CommitHash: "abc123",
	}
	gitCacheErr = errGitUnavailable
	gitCacheTime = time.Now()

	// Clear cache
	ClearGitCache()

	// Verify cache is cleared
	if gitCache != nil {
		t.Error("Expected gitCache to be nil after clearing")
	}
	if gitCacheErr != nil {
		t.Error("Expected gitCacheErr to be nil after clearing")
	}
	if !gitCacheTime.IsZero() {
		t.Error("Expected gitCacheTime to be zero after clearing")
	}
}
