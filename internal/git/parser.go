package git

import (
	"fmt"
	"regexp"
	"strings"
)

// RepoRef represents a parsed git repository reference
type RepoRef struct {
	Host  string // github.com
	Owner string // user
	Repo  string // repo
	Path  string // apps/blog (optional subfolder)
	Ref   string // v1.0.0, main, abc1234 (optional)
}

// FullURL returns the HTTPS clone URL for the repository
func (r *RepoRef) FullURL() string {
	return fmt.Sprintf("https://%s/%s/%s.git", r.Host, r.Owner, r.Repo)
}

// String returns a human-readable representation
func (r *RepoRef) String() string {
	s := fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
	if r.Path != "" {
		s += "/" + r.Path
	}
	if r.Ref != "" {
		s += "@" + r.Ref
	}
	return s
}

// ParseURL parses various git URL formats into a RepoRef
// Supported formats:
//   - github.com/user/repo
//   - github.com/user/repo/path/to/app
//   - github.com/user/repo@v1.0.0
//   - github.com/user/repo/path/to/app@v1.0.0
//   - github:user/repo (shorthand)
//   - github:user/repo/app@main
//   - https://github.com/user/repo
//   - https://github.com/user/repo/tree/main/path
func ParseURL(input string) (*RepoRef, error) {
	input = strings.TrimSpace(input)

	// Handle shorthand: github:user/repo -> github.com/user/repo
	if strings.HasPrefix(input, "github:") {
		input = "github.com/" + strings.TrimPrefix(input, "github:")
	}

	// Strip https:// or http://
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")

	// Strip .git suffix
	input = strings.TrimSuffix(input, ".git")

	// Handle GitHub tree URLs: github.com/user/repo/tree/branch/path
	treeRe := regexp.MustCompile(`^(github\.com)/([^/]+)/([^/]+)/tree/([^/]+)(?:/(.*))?$`)
	if matches := treeRe.FindStringSubmatch(input); matches != nil {
		return &RepoRef{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
			Ref:   matches[4],
			Path:  matches[5],
		}, nil
	}

	// Handle GitHub blob URLs: github.com/user/repo/blob/branch/path (treat as tree)
	blobRe := regexp.MustCompile(`^(github\.com)/([^/]+)/([^/]+)/blob/([^/]+)(?:/(.*))?$`)
	if matches := blobRe.FindStringSubmatch(input); matches != nil {
		return &RepoRef{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
			Ref:   matches[4],
			Path:  matches[5],
		}, nil
	}

	// Extract @ref suffix
	var ref string
	if idx := strings.LastIndex(input, "@"); idx != -1 {
		ref = input[idx+1:]
		input = input[:idx]
	}

	// Parse remaining path: host/owner/repo[/path...]
	parts := strings.Split(input, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid URL format: need at least host/owner/repo")
	}

	host := parts[0]
	owner := parts[1]
	repo := parts[2]
	var path string

	if len(parts) > 3 {
		path = strings.Join(parts[3:], "/")
	}

	// Validate host
	if !isValidHost(host) {
		return nil, fmt.Errorf("unsupported host: %s (only github.com supported)", host)
	}

	return &RepoRef{
		Host:  host,
		Owner: owner,
		Repo:  repo,
		Path:  path,
		Ref:   ref,
	}, nil
}

// isValidHost checks if the host is supported
func isValidHost(host string) bool {
	// Currently only GitHub is supported
	return host == "github.com"
}
