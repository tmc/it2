# it2

Command it2 provides comprehensive control over iTerm2 through its WebSocket API.

The it2 command is a powerful CLI tool that enables terminal automation, session management, and access to advanced iTerm2 features. It's designed for both interactive use and scripting workflows.
## Usage

	it2 [global-flags] <command> [command-flags] [arguments]

## Global Flags

	--format string      Output format: text, json, yaml (default "text")
	--timeout duration   Connection timeout (default 5s)
	--url string         WebSocket URL (default "ws://localhost:1912")

## Installation

Install the latest version:

	go install github.com/tmc/it2/cmd/it2@latest

Or download pre-built binaries from GitHub releases.

## Quick Start

Basic session management:

	# List all terminal sessions
	it2 session list

	# Send text to current session
	it2 session send-text "echo Hello!"

	# Create a new tab
	it2 tab create

	# Split current pane
	it2 session split --vertical

## Command Categories

Application Control:

  - app: Control iTerm2 application (focus, variables, properties)
  - auth: Manage API authentication and connection
  - arrangement: Save and restore window arrangements

Session & Tab Management:

  - session: Create, list, close, split, activate, restart sessions
  - tab: Create, list, close, move, activate tabs
  - window: Create, list, close, focus windows

Text & Content Operations:

  - text: Buffer operations, cursor control, search, selection
  - selection: Control and retrieve text selection
  - annotation: Add and manage session annotations
  - badge: Display informational badges on sessions

Shell Integration Features (requires Shell Integration):

  - prompt: Access command history, prompts, exit codes
  - job: Monitor running processes and jobs

Customization:

  - profile: Manage terminal profiles and their properties
  - color: Import/export color presets and themes
  - variable: Get/set iTerm2 variables with scope control

Monitoring & Events:

  - notification: Subscribe to real-time iTerm2 events
  - broadcast: Manage broadcast input domains
  - statusbar: Configure status bar components

Advanced Features:

  - tmux: Control tmux integration

## Configuration

Global flags available for all commands:

	--format string      Output format: text, json, yaml (default "text")
	--timeout duration   Connection timeout (default 5s)
	--url string         WebSocket URL (default "ws://localhost:1912")

## Environment Variables

Key environment variables that affect behavior:

	ITERM_SESSION_ID     Current session ID (automatically set by iTerm2)
	ITERM2_COOKIE        Authentication cookie (auto-requested if not set)
	ITERM2_KEY           Authentication key (auto-requested if not set)
	ITERM2_DEBUG         Enable debug logging (set to "1")

## Authentication

The it2 command automatically handles authentication with iTerm2:

 1. On first connection, it requests credentials from iTerm2 via AppleScript
 2. iTerm2 prompts the user to allow API access
 3. Credentials are cached for the session duration
 4. No manual authentication setup is required

Check authentication status:

	it2 auth check

Force re-authentication:

	it2 auth request

## Session Identification

Many commands work with session IDs in two formats:

  - Full format: w0t1p12:C3D91F33-3805-47E2-A3F6-B8AED6EC2209
  - UUID only: C3D91F33-3805-47E2-A3F6-B8AED6EC2209

Commands automatically normalize session IDs and fall back to $ITERM\_SESSION\_ID when appropriate, making current session operations seamless.

## Output Formats

Most commands support multiple output formats:

	# Human-readable text (default)
	it2 session list

	# JSON for scripting
	it2 session list --format json

	# YAML for configuration
	it2 profile get "Default" --format yaml

## Shell Integration

Advanced features require iTerm2's Shell Integration:

  - Command history access (prompt commands)
  - Exit code tracking
  - Job monitoring
  - Directory tracking

Enable in iTerm2: Menu → iTerm2 → Install Shell Integration

Example Shell Integration workflows:

	# Show command history with exit codes
	it2 prompt list

	# Search for failed commands
	it2 prompt search "git" | grep "Exit: [^0]"

	# Monitor running jobs
	it2 job list <session-id>

## Real-time Monitoring

Subscribe to live iTerm2 events:

	# Monitor keystrokes
	it2 notification monitor --type keystroke

	# Watch for new sessions
	it2 notification monitor --type session

	# Monitor variable changes
	it2 variable monitor session user.myvar

## Scripting Examples

Useful patterns for automation:

	# Create development environment
	TAB1=$(it2 tab create --format json | jq -r '.tab_id')
	it2 session send-text "cd ~/project && vim"
	it2 session split --horizontal
	it2 session send-text "npm run dev"

	# Broadcast command to all sessions
	for session in $(it2 session list --format json | jq -r '.[].id'); do
	    it2 session send-text "$session" "git pull"
	done

	# Save terminal state
	it2 arrangement save "development-setup"

## Common Use Cases

Development workflow automation:

	# Setup development environment
	it2 tab create "Development"
	it2 session send-text "cd ~/project"
	it2 session split --vertical
	it2 session send-text "npm run dev"
	it2 session split --horizontal
	it2 session send-text "git status"

Remote server management:

	# Connect to multiple servers
	for server in web1 web2 db1; do
	    it2 tab create "$server"
	    it2 session send-text "ssh $server"
	done

Monitoring and logging:

	# Watch logs in real-time
	it2 session send-text "tail -f /var/log/app.log"
	it2 notification monitor --type screen | grep "ERROR"

Session backup and restore:

	# Save current arrangement
	it2 arrangement save "$(date +%Y%m%d_%H%M%S)"

	# List saved arrangements
	it2 arrangement list

	# Restore arrangement
	it2 arrangement restore "20240101_120000"

## Troubleshooting

Common issues and solutions:

Connection problems:

	# Check if iTerm2 API is enabled
	it2 auth check

	# Enable debug output
	ITERM2_DEBUG=1 it2 session list

Shell Integration not working:

	# Verify Shell Integration is installed
	echo $ITERM_SESSION_ID  # Should show session ID

	# Reinstall if needed
	curl -L https://iterm2.com/shell_integration/install_shell_integration_and_utilities.sh | bash

## Requirements

  - iTerm2 version 3.3.0 or later
  - Python API enabled: Preferences → General → Magic → Enable Python API
  - macOS (iTerm2 is macOS-only)
  - Shell Integration (for prompt/job commands)

## Exit Codes

The it2 command uses standard exit codes:

	0    Success
	1    General error (authentication, connection, invalid arguments)
	2    Command usage error (wrong number of arguments, invalid flags)

## Files

iTerm2 uses these files for configuration and communication:

	~/Library/Application Support/iTerm2/private/socket
	    Unix domain socket for local API communication (preferred)

	ws://localhost:1912
	    WebSocket endpoint for API communication (fallback)

## Environment Variables

The following environment variables affect it2 behavior:

	ITERM_SESSION_ID
	    Current session ID (automatically set by iTerm2 with Shell Integration)
	    Format: w0t1p12:C3D91F33-3805-47E2-A3F6-B8AED6EC2209

	ITERM2_COOKIE
	    Authentication cookie (auto-requested if not set)

	ITERM2_KEY
	    Authentication key (auto-requested if not set)

	ITERM2_DEBUG
	    Enable debug logging when set to "1"
	    Shows WebSocket messages and connection details

## More Information

Documentation and examples:

  - Project repository: [https://github.com/tmc/it2](https://github.com/tmc/it2)
  - iTerm2 API documentation: [https://iterm2.com/documentation-api.html](https://iterm2.com/documentation-api.html)
  - Shell Integration guide: [https://iterm2.com/documentation-shell-integration.html](https://iterm2.com/documentation-shell-integration.html)

Get help for any command:

	it2 help [command]
	it2 [command] --help
	it2 [command] [subcommand] --help
