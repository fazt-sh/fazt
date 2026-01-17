package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneOptions configures a clone operation
type CloneOptions struct {
	URL       string // Full HTTPS URL (from RepoRef.FullURL())
	Path      string // Subfolder within repo (optional)
	Ref       string // Tag, branch, or commit (optional, defaults to HEAD)
	TargetDir string // Where to clone
}

// CloneResult contains metadata about a successful clone
type CloneResult struct {
	CommitSHA   string    // Full commit SHA
	CommitTime  time.Time // Commit timestamp
	RefResolved string    // What ref resolved to
	Files       int       // Number of files copied
}

// Clone fetches a repo (or subfolder) to local directory
func Clone(opts CloneOptions) (*CloneResult, error) {
	// Clone to temp directory first
	tmpDir, err := os.MkdirTemp("", "fazt-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone options
	cloneOpts := &git.CloneOptions{
		URL:      opts.URL,
		Progress: io.Discard,
		Depth:    1, // Shallow clone for efficiency
	}

	// If a specific ref is requested
	if opts.Ref != "" {
		// Try as branch/tag first
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Ref)
		cloneOpts.SingleBranch = true
	}

	// Clone
	repo, err := git.PlainClone(tmpDir, false, cloneOpts)
	if err != nil {
		// If branch didn't work, try as tag
		if opts.Ref != "" {
			cloneOpts.ReferenceName = plumbing.NewTagReferenceName(opts.Ref)
			repo, err = git.PlainClone(tmpDir, false, cloneOpts)
		}
		if err != nil {
			// Try full clone without ref constraint (for commit SHAs)
			cloneOpts.ReferenceName = ""
			cloneOpts.SingleBranch = false
			repo, err = git.PlainClone(tmpDir, false, cloneOpts)
		}
		if err != nil {
			return nil, fmt.Errorf("clone failed: %w", err)
		}
	}

	// If a specific commit was requested, checkout
	if opts.Ref != "" && len(opts.Ref) >= 7 {
		// Might be a commit SHA
		w, err := repo.Worktree()
		if err == nil {
			hash := plumbing.NewHash(opts.Ref)
			if hash.IsZero() == false {
				w.Checkout(&git.CheckoutOptions{Hash: hash})
			}
		}
	}

	// Get HEAD commit info
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	// Determine source directory (repo root or subfolder)
	sourceDir := tmpDir
	if opts.Path != "" {
		sourceDir = filepath.Join(tmpDir, opts.Path)
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("path not found in repo: %s", opts.Path)
		}
	}

	// Copy files to target directory
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target dir: %w", err)
	}

	fileCount, err := copyDir(sourceDir, opts.TargetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to copy files: %w", err)
	}

	return &CloneResult{
		CommitSHA:   head.Hash().String(),
		CommitTime:  commit.Author.When,
		RefResolved: head.Name().Short(),
		Files:       fileCount,
	}, nil
}

// GetLatestCommit returns the HEAD commit SHA for a repo
func GetLatestCommit(repoURL string, ref string) (string, error) {
	// Use ls-remote equivalent
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list remote refs: %w", err)
	}

	// Default to HEAD
	if ref == "" {
		ref = "HEAD"
	}

	for _, r := range refs {
		name := r.Name().String()
		short := r.Name().Short()

		// Match HEAD
		if ref == "HEAD" && name == "HEAD" {
			return r.Hash().String(), nil
		}

		// Match branch or tag
		if short == ref {
			return r.Hash().String(), nil
		}

		// Match refs/heads/xxx or refs/tags/xxx
		if name == "refs/heads/"+ref || name == "refs/tags/"+ref {
			return r.Hash().String(), nil
		}
	}

	// If ref looks like a commit SHA, return it directly
	if len(ref) >= 7 && len(ref) <= 40 {
		return ref, nil
	}

	return "", fmt.Errorf("ref not found: %s", ref)
}

// PrebuiltBranches are common branch names for pre-built output
var PrebuiltBranches = []string{
	"fazt-dist", // Recommended convention
	"dist",
	"release",
	"gh-pages",
}

// FindPrebuiltBranch checks if a repo has a pre-built branch
// Returns the branch name if found, empty string otherwise
func FindPrebuiltBranch(repoURL string) string {
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return ""
	}

	// Build a set of available branches
	branches := make(map[string]bool)
	for _, r := range refs {
		if r.Name().IsBranch() {
			branches[r.Name().Short()] = true
		}
	}

	// Check for pre-built branches in priority order
	for _, branch := range PrebuiltBranches {
		if branches[branch] {
			return branch
		}
	}

	return ""
}

// copyDir copies a directory recursively, excluding .git
func copyDir(src, dst string) (int, error) {
	fileCount := 0

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		fileCount++
		return os.Chmod(targetPath, info.Mode())
	})

	return fileCount, err
}
