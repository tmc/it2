package setup

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/auth"
	"github.com/tmc/it2/internal/client"
	"github.com/tmc/it2/internal/launcher"
)

// NewCommand creates the setup command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Bootstrap iTerm2 and it2 for first use",
		Long: `Set up iTerm2 and it2 from scratch.

This command walks through the full bootstrap:
  1. Checks if iTerm2 is installed (installs via Homebrew if not)
  2. Launches iTerm2
  3. Enables the iTerm2 API
  4. Requests API authentication
  5. Verifies the connection works
  6. Shows next steps`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
	return cmd
}

func runSetup() error {
	fmt.Fprintln(os.Stderr, "it2 setup")
	fmt.Fprintln(os.Stderr, "=========")
	fmt.Fprintln(os.Stderr)

	// Step 1: Check / install iTerm2
	fmt.Fprintln(os.Stderr, "[1/5] Checking iTerm2 installation...")
	if launcher.IsITerm2Installed() {
		fmt.Fprintln(os.Stderr, "      iTerm2 is already installed.")
	} else {
		fmt.Fprintln(os.Stderr, "      iTerm2 is not installed. Attempting install...")
		if err := launcher.InstallITerm2(); err != nil {
			return fmt.Errorf("step 1 failed: %w", err)
		}
	}

	// Step 2: Launch iTerm2
	fmt.Fprintln(os.Stderr, "[2/5] Launching iTerm2...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if launcher.IsITerm2Running() {
		fmt.Fprintln(os.Stderr, "      iTerm2 is already running.")
	} else {
		if err := launcher.LaunchITerm2(ctx); err != nil {
			return fmt.Errorf("step 2 failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "      iTerm2 launched.")
	}

	// Step 3: Enable API automation
	fmt.Fprintln(os.Stderr, "[3/5] Enabling iTerm2 API automation...")
	if err := auth.CheckAutomationEnabled(); err != nil {
		if err := auth.EnableAutomation(); err != nil {
			return fmt.Errorf("step 3 failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "      API automation enabled.")
	} else {
		fmt.Fprintln(os.Stderr, "      API automation is already enabled.")
	}

	// Step 4: Request authentication
	fmt.Fprintln(os.Stderr, "[4/5] Requesting API authentication...")
	if auth.HasAuthentication() {
		fmt.Fprintln(os.Stderr, "      Authentication is already configured.")
	} else {
		if err := auth.RequestAuthentication(); err != nil {
			fmt.Fprintln(os.Stderr, "      Warning: could not get authentication automatically.")
			fmt.Fprintln(os.Stderr, "      iTerm2 may prompt you to allow API access.")
			fmt.Fprintf(os.Stderr, "      (%v)\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "      Authentication configured.")
		}
	}

	// Step 5: Test connection
	fmt.Fprintln(os.Stderr, "[5/5] Testing connection...")
	c := client.New("ws://localhost:1912")
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()
	if err := c.Connect(testCtx); err != nil {
		fmt.Fprintln(os.Stderr, "      Warning: could not connect to iTerm2 API.")
		fmt.Fprintf(os.Stderr, "      (%v)\n", err)
		fmt.Fprintln(os.Stderr, "      You may need to restart iTerm2 for API changes to take effect.")
	} else {
		c.Close()
		fmt.Fprintln(os.Stderr, "      Connection successful!")
	}

	// Welcome message
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Setup complete! Get started with:")
	fmt.Fprintln(os.Stderr, "  it2 quickstart        Full guide and examples")
	fmt.Fprintln(os.Stderr, "  it2 session list      List current sessions")
	fmt.Fprintln(os.Stderr, "  it2 --help            See all commands")
	return nil
}
