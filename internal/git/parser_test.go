package git

import (
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RepoRef
		wantErr bool
	}{
		{
			name:  "simple github URL",
			input: "github.com/user/repo",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "github URL with path",
			input: "github.com/user/repo/apps/blog",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Path: "apps/blog"},
		},
		{
			name:  "github URL with ref",
			input: "github.com/user/repo@v1.0.0",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Ref: "v1.0.0"},
		},
		{
			name:  "github URL with path and ref",
			input: "github.com/user/repo/app@main",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Path: "app", Ref: "main"},
		},
		{
			name:  "github shorthand",
			input: "github:user/repo",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "github shorthand with path and ref",
			input: "github:user/repo/apps/blog@v2.0.0",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Path: "apps/blog", Ref: "v2.0.0"},
		},
		{
			name:  "https URL",
			input: "https://github.com/user/repo",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "URL with .git suffix",
			input: "github.com/user/repo.git",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
		},
		{
			name:  "github tree URL",
			input: "https://github.com/user/repo/tree/main/apps/blog",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Ref: "main", Path: "apps/blog"},
		},
		{
			name:  "github tree URL without path",
			input: "https://github.com/user/repo/tree/develop",
			want:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Ref: "develop"},
		},
		{
			name:    "invalid - too few parts",
			input:   "github.com/user",
			wantErr: true,
		},
		{
			name:    "invalid - unsupported host",
			input:   "gitlab.com/user/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %v, want %v", got.Host, tt.want.Host)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %v, want %v", got.Owner, tt.want.Owner)
			}
			if got.Repo != tt.want.Repo {
				t.Errorf("Repo = %v, want %v", got.Repo, tt.want.Repo)
			}
			if got.Path != tt.want.Path {
				t.Errorf("Path = %v, want %v", got.Path, tt.want.Path)
			}
			if got.Ref != tt.want.Ref {
				t.Errorf("Ref = %v, want %v", got.Ref, tt.want.Ref)
			}
		})
	}
}

func TestRepoRef_FullURL(t *testing.T) {
	ref := &RepoRef{
		Host:  "github.com",
		Owner: "user",
		Repo:  "repo",
	}
	want := "https://github.com/user/repo.git"
	if got := ref.FullURL(); got != want {
		t.Errorf("FullURL() = %v, want %v", got, want)
	}
}

func TestRepoRef_String(t *testing.T) {
	tests := []struct {
		ref  RepoRef
		want string
	}{
		{
			ref:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo"},
			want: "github.com/user/repo",
		},
		{
			ref:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Path: "apps/blog"},
			want: "github.com/user/repo/apps/blog",
		},
		{
			ref:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Ref: "v1.0.0"},
			want: "github.com/user/repo@v1.0.0",
		},
		{
			ref:  RepoRef{Host: "github.com", Owner: "user", Repo: "repo", Path: "app", Ref: "main"},
			want: "github.com/user/repo/app@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ref.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
