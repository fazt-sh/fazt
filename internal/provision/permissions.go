package provision

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallBinary copies the current executable to the target path
func InstallBinary(targetPath string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Check if we are already running from the target location
	currentAbs, err := filepath.Abs(currentExe)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get target absolute path: %w", err)
	}

	if currentAbs == targetAbs {
		fmt.Printf("Binary already at target location: %s\n", targetPath)
		return nil
	}

	fmt.Printf("Installing binary from %s to %s...\n", currentExe, targetPath)

	src, err := os.Open(currentExe)
	if err != nil {
		return fmt.Errorf("failed to open source binary: %w", err)
	}
	defer src.Close()

	// Create or truncate target
	dst, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target binary: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	return nil
}

// SetCapabilities sets CAP_NET_BIND_SERVICE on the binary
func SetCapabilities(binaryPath string) error {
	path, err := exec.LookPath("setcap")
	if err != nil {
		fmt.Println("Warning: 'setcap' not found. If you want to run on port 80/443 without root, install libcap2-bin.")
		return nil
	}

	fmt.Printf("Setting capabilities on %s...\n", binaryPath)
	cmd := exec.Command(path, "CAP_NET_BIND_SERVICE=+eip", binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set capabilities: %w\nOutput: %s", err, string(output))
	}

	return nil
}
