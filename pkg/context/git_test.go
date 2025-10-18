package context

import (
	"testing"
	"time"
)

func TestConvertSSHToHTTPS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "github ssh format",
			input: "git@github.com:user/repo.git",
			want:  "https://github.com/user/repo",
		},
		{
			name:  "bitbucket ssh format",
			input: "ssh://git@bitbucket.org/user/repo.git",
			want:  "https://bitbucket.org/user/repo",
		},
		{
			name:  "already https",
			input: "https://github.com/user/repo.git",
			want:  "https://github.com/user/repo",
		},
		{
			name:  "no git suffix",
			input: "https://github.com/user/repo",
			want:  "https://github.com/user/repo",
		},
		{
			name:  "gitlab ssh format",
			input: "git@gitlab.com:user/repo.git",
			want:  "https://gitlab.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertSSHToHTTPS(tt.input)
			if got != tt.want {
				t.Errorf("convertSSHToHTTPS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClearGitCache(t *testing.T) {
	// Set up cache
	gitCache = &GitInfo{
		RepoURL:    "https://github.com/test/repo",
		CommitHash: "abc123",
	}
	gitCacheTime = time.Now()

	// Clear cache
	ClearGitCache()

	// Verify cache is cleared
	if gitCache != nil {
		t.Error("Expected gitCache to be nil after clearing")
	}
	if !gitCacheTime.IsZero() {
		t.Error("Expected gitCacheTime to be zero after clearing")
	}
}
