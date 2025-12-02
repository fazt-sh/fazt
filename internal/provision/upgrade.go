package provision

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoOwner = "fazt-sh"
	repoName  = "fazt"
)

// ReleaseInfo represents GitHub release data
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Upgrade checks for updates and upgrades the binary
func Upgrade(currentVersion string) error {
	fmt.Println("Checking for updates...")

	// 1. Fetch latest release info
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}

	if release.TagName == currentVersion {
		fmt.Printf("You are already using the latest version (%s)\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: %s (current: %s)\n", release.TagName, currentVersion)
	fmt.Print("Do you want to upgrade? [y/N] ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Upgrade cancelled.")
		return nil
	}

	// 2. Find matching asset
	assetURL, assetName, err := findAssetURL(release)
	if err != nil {
		return fmt.Errorf("failed to find compatible binary: %w", err)
	}

	// 3. Download and extract
	fmt.Printf("Downloading %s...\n", assetName)
	tmpDir, err := os.MkdirTemp("", "fazt-upgrade")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := downloadAndExtract(assetURL, tmpDir); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// 4. Replace binary
	newBinary := filepath.Join(tmpDir, "fazt")
	// Fallback check if it has extension or different name
	if _, err := os.Stat(newBinary); os.IsNotExist(err) {
		entries, _ := os.ReadDir(tmpDir)
		for _, e := range entries {
			if !e.IsDir() {
				newBinary = filepath.Join(tmpDir, e.Name())
				break
			}
		}
	}

	currentBinary, err := os.Executable()
	if err != nil {
		return err
	}

	// Resolve symlinks
	currentBinary, err = filepath.EvalSymlinks(currentBinary)
	if err != nil {
		return err
	}

	fmt.Printf("Upgrading binary at: %s\n", currentBinary)

	// Move current binary to backup
	backupPath := currentBinary + ".old"
	if err := os.Rename(currentBinary, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to current location
	if err := moveFile(newBinary, currentBinary); err != nil {
		os.Rename(backupPath, currentBinary) // Restore
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	if err := os.Chmod(currentBinary, 0755); err != nil {
		fmt.Printf("Warning: failed to set executable permissions: %v\n", err)
	}

	fmt.Println("✓ Binary updated successfully")

	// 5. Apply permissions (Linux) if we are in a system path or user asks
	// Only apply setcap if we are on Linux
	if runtime.GOOS == "linux" {
		// Check if we can/should use setcap
		// Simple heuristic: if we are root, try it.
		if os.Geteuid() == 0 {
			fmt.Println("Applying capabilities (setcap)...")
			cmd := exec.Command("setcap", "CAP_NET_BIND_SERVICE=+eip", currentBinary)
			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Warning: setcap failed: %v\n%s\n", err, out)
			}
		}
	}

	// 6. Restart service
	// Only try to restart if systemd service exists and is active
	if isServiceActive("fazt") {
		fmt.Println("Restarting systemd service...")
		if err := Systemctl("restart", "fazt"); err != nil {
			fmt.Printf("Warning: Service restart failed: %v\n", err)
		} else {
			fmt.Println("✓ Service restarted")
		}
	} else {
		fmt.Println("Note: 'fazt' systemd service not detected or active. Please restart your server manually if needed.")
	}

	return nil
}

func isServiceActive(name string) bool {
	cmd := exec.Command("systemctl", "is-active", name)
	return cmd.Run() == nil
}

func getLatestRelease() (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func findAssetURL(release *ReleaseInfo) (string, string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Normalize patterns
	// Our release workflow uses: fazt-v0.2.0-linux-amd64.tar.gz
	pattern := fmt.Sprintf("%s-%s", osName, arch)

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, pattern) && strings.HasSuffix(asset.Name, ".tar.gz") {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("no asset found for %s/%s", osName, arch)
}

func downloadAndExtract(url, destDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle tar.gz
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// We are looking for the binary. It might be named "fazt" or "fazt-linux-amd64" depending on build.
		// In our workflow: go build -o fazt-linux-amd64
		// So we extract whatever is there to destDir.
		
		target := filepath.Join(destDir, header.Name)
		
		switch header.Typeflag {
		case tar.TypeReg:
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func moveFile(src, dst string) error {
	// Try atomic rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback to copy (e.g. across partitions)
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, input, 0755); err != nil {
		return err
	}
	return os.Remove(src)
}
