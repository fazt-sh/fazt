package provision

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fazt-sh/fazt/internal/term"
)

// ConfigureFirewall attempts to configure UFW if it exists
func ConfigureFirewall() error {
	path, err := exec.LookPath("ufw")
	if err != nil {
		term.Info("UFW not found, skipping firewall configuration")
		return nil
	}

	term.Step("Configuring Firewall (UFW)...")

	// Allow SSH to prevent lockout (just in case)
	// We don't want to enable UFW if it's disabled, as that's a big policy change.
	// We only want to allow ports if UFW is *present*.

	rules := []string{
		"allow 80/tcp",
		"allow 443/tcp",
		"allow ssh", // Safety first!
	}

	for _, rule := range rules {
		cmd := exec.Command(path, rule)
		// UFW output can be noisy, so we might want to capture it or silence it
		// But let's show it for transparency
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		fmt.Printf("Running: ufw %s\n", rule)
		if err := cmd.Run(); err != nil {
			term.Warn("Failed to run 'ufw %s': %v", rule, err)
		}
	}
	
	// Reload is good practice
	exec.Command(path, "reload").Run()

	term.Success("Firewall rules updated")
	return nil
}
