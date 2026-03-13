package git

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

var (
	ErrInvalidRemoteURL    = errors.New("invalid remote URL")
	ErrUnsupportedProtocol = errors.New("unsupported protocol")
)

type RemoteInfo struct {
	Host  string
	Owner string
	Repo  string
}

type RemoteURLResolver interface {
	RemoteURL(remoteName string) (string, error)
}

type ExecRemoteURLResolver struct {
	dir string
}

func NewExecRemoteURLResolver(dir string) *ExecRemoteURLResolver {
	return &ExecRemoteURLResolver{dir: dir}
}

func (r *ExecRemoteURLResolver) RemoteURL(remoteName string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	if r.dir != "" {
		cmd.Dir = r.dir
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL for %q: %w", remoteName, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func ParseRemoteURL(rawURL string) (*RemoteInfo, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, ErrInvalidRemoteURL
	}

	if strings.Contains(rawURL, "://") {
		return parseSchemeURL(rawURL)
	}

	if strings.Contains(rawURL, ":") {
		return parseSSHURL(rawURL)
	}

	return nil, ErrInvalidRemoteURL
}

func parseSchemeURL(rawURL string) (*RemoteInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, ErrInvalidRemoteURL
	}

	switch u.Scheme {
	case "http", "https":
	case "file":
		return nil, ErrUnsupportedProtocol
	default:
		return nil, ErrUnsupportedProtocol
	}

	if u.Host == "" {
		return nil, ErrInvalidRemoteURL
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return nil, ErrInvalidRemoteURL
	}

	owner, repo, ok := splitOwnerRepo(path)
	if !ok {
		return nil, ErrInvalidRemoteURL
	}

	return &RemoteInfo{
		Host:  u.Host,
		Owner: owner,
		Repo:  repo,
	}, nil
}

func parseSSHURL(rawURL string) (*RemoteInfo, error) {
	// git@host:owner/repo.git
	colonIdx := strings.Index(rawURL, ":")
	host := rawURL[:colonIdx]
	if atIdx := strings.Index(host, "@"); atIdx >= 0 {
		host = host[atIdx+1:]
	}

	if host == "" {
		return nil, ErrInvalidRemoteURL
	}

	path := strings.Trim(rawURL[colonIdx+1:], "/")
	if path == "" {
		return nil, ErrInvalidRemoteURL
	}

	owner, repo, ok := splitOwnerRepo(path)
	if !ok {
		return nil, ErrInvalidRemoteURL
	}

	return &RemoteInfo{
		Host:  host,
		Owner: owner,
		Repo:  repo,
	}, nil
}

func splitOwnerRepo(path string) (owner, repo string, ok bool) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", false
	}

	repo = parts[len(parts)-1]
	repo = strings.TrimSuffix(repo, ".git")
	if repo == "" {
		return "", "", false
	}

	owner = strings.Join(parts[:len(parts)-1], "/")
	if owner == "" {
		return "", "", false
	}

	return owner, repo, true
}
