The installation script issue has been resolved.

**Problem:**
When running `curl ... | bash`, the script's standard input (stdin) is connected to the curl output (the script content itself). The `read` commands in the script were trying to read from stdin, which was already exhausted or reserved for the script execution, causing the prompts to be skipped immediately.

**Fix:**
I have updated both `install.sh` and `docs/install.sh` to explicitly redirect input from `/dev/tty` for all interactive prompts (e.g., `read ... < /dev/tty`). This forces the script to wait for user input from the terminal even when piped.

**Verification:**
- Verified that `fazt service install` (the Go command called by the script) is non-interactive and relies on the script to gather arguments, which confirms the script's role is critical.
- Performed syntax checks (`bash -n`) on the modified scripts.
- Also fixed a `.gitignore` misconfiguration that was incorrectly ignoring the `cmd/server` directory.

The install command:
`curl -s https://fazt-sh.github.io/fazt/install.sh | bash`
Should now correctly prompt the user for installation type and details.