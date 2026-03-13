package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *RemoteInfo
		wantErr error
	}{
		{
			name:  "HTTPS with .git suffix",
			input: "https://github.com/user/repo.git",
			want:  &RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "HTTPS without .git suffix",
			input: "https://github.com/user/repo",
			want:  &RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "SSH with .git suffix",
			input: "git@github.com:user/repo.git",
			want:  &RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "SSH without .git suffix",
			input: "git@github.com:user/repo",
			want:  &RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "GitLab nested groups",
			input: "https://gitlab.com/group/subgroup/repo.git",
			want:  &RemoteInfo{Host: "gitlab.com", Owner: "group/subgroup", Repo: "repo"},
		},
		{
			name:  "custom host SSH",
			input: "git@git.company.com:team/project.git",
			want:  &RemoteInfo{Host: "git.company.com", Owner: "team", Repo: "project"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrInvalidRemoteURL,
		},
		{
			name:    "malformed URL",
			input:   "not-a-url",
			wantErr: ErrInvalidRemoteURL,
		},
		{
			name:    "file protocol",
			input:   "file:///path/to/repo",
			wantErr: ErrUnsupportedProtocol,
		},
		{
			name:  "trailing slash",
			input: "https://github.com/user/repo/",
			want:  &RemoteInfo{Host: "github.com", Owner: "user", Repo: "repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRemoteURL(tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
