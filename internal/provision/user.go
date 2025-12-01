package provision

import (
	"fmt"
	"os/exec"
	"os/user"
)

// EnsureUser ensures the specified system user exists.
// If not, it creates it with a home directory.
func EnsureUser(username string) error {
	// Check if user exists
	_, err := user.Lookup(username)
	if err == nil {
		return nil // User exists
	}

	// UnknownUserError is returned if user doesn't exist
	if _, ok := err.(user.UnknownUserError); !ok {
		return fmt.Errorf("failed to lookup user %s: %w", username, err)
	}

	fmt.Printf("Creating system user '%s'...\n", username)

	// Create user
	// -m: Create home directory
	// -s /bin/false: No shell access (security)
	// -U: Create a group with the same name
	cmd := exec.Command("useradd", "-m", "-s", "/bin/false", "-U", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w\nOutput: %s", username, err, string(output))
	}

	return nil
}

// GetUserUIDGID returns the UID and GID for a username
func GetUserUIDGID(username string) (string, string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", "", err
	}
	return u.Uid, u.Gid, nil
}
