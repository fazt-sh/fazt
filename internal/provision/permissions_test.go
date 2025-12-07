package provision

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallBinary_SelfCopy(t *testing.T) {
	// 1. Create a dummy binary (just a text file is fine for this test, as we are mocking os.Executable somewhat or just using a file)
	// Actually, InstallBinary calls os.Executable() which returns the *current test binary* path.
	// So we can point targetPath to os.Executable() and see if it fails or succeeds.
	
	currentExe, err := os.Executable()
	if err != nil {
		t.Fatalf("failed to get executable: %v", err)
	}
	
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "fazt-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Case 1: Target is same as Current (Should succeed and do nothing)
	// Note: We can't really verify it "did nothing" easily without mocking logs, 
	// but we can verify it doesn't return error "text file busy".
	err = InstallBinary(currentExe)
	if err != nil {
		t.Errorf("InstallBinary(currentExe) failed: %v", err)
	}
	
	// Case 2: Target is different (Should copy)
	targetPath := filepath.Join(tmpDir, "fazt-copy")
	err = InstallBinary(targetPath)
	if err != nil {
		t.Errorf("InstallBinary(targetPath) failed: %v", err)
	}
	
	// Verify copy exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Errorf("Target file not created at %s", targetPath)
	}
	
	// Verify content size matches roughly (optional)
	srcInfo, _ := os.Stat(currentExe)
	dstInfo, _ := os.Stat(targetPath)
	if srcInfo.Size() != dstInfo.Size() {
		t.Errorf("Size mismatch: src=%d, dst=%d", srcInfo.Size(), dstInfo.Size())
	}
}
