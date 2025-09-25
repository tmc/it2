# it2 Configuration Guide

Complete guide to configuring the it2 CLI tool for optimal usage.

## Configuration Methods

it2 supports configuration through multiple methods, in order of precedence:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Built-in defaults** (lowest priority)

## Configuration File

### Location

Default configuration file location:
- **macOS/Linux**: `~/.it2/config.yaml`
- **Custom location**: Use `IT2_CONFIG` environment variable

### File Format

Configuration uses YAML format:

```yaml
# ~/.it2/config.yaml

# Connection settings
url: "ws://localhost:1912"
timeout: "10s"

# Output settings
format: "table"
color: true
template: ""

# Retry settings
retry_attempts: 3
retry_delay: "1s"

# Logging settings
log_level: "info"
verbose: false
debug: false

# Progress indicators
show_progress: true

# Authentication (usually auto-managed)
cookie: ""
key: ""
```

### Managing Configuration

**View current configuration:**
```bash
it2 config show
```

**Show configuration file path:**
```bash
it2 config path
```

**Create default configuration file:**
```bash
it2 config init
```

## Configuration Options

### Connection Settings

#### `url`
- **Type**: String
- **Default**: `ws://localhost:1912`
- **Description**: WebSocket URL for iTerm2 API
- **Environment**: `IT2_URL`
- **Example**:
  ```yaml
  url: "ws://192.168.1.100:1912"  # Remote iTerm2
  ```

#### `timeout`
- **Type**: Duration string
- **Default**: `5s`
- **Description**: Connection timeout for API calls
- **Environment**: `IT2_TIMEOUT`
- **Examples**:
  ```yaml
  timeout: "30s"    # 30 seconds
  timeout: "2m"     # 2 minutes
  timeout: "500ms"  # 500 milliseconds
  ```

### Output Settings

#### `format`
- **Type**: String
- **Default**: `table`
- **Options**: `table`, `text`, `json`, `yaml`
- **Description**: Default output format for commands
- **Environment**: `IT2_FORMAT`
- **Examples**:
  ```yaml
  format: "json"  # JSON output by default
  ```

#### `color`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Enable colored output
- **Environment**: `IT2_COLOR`
- **Examples**:
  ```yaml
  color: false  # Disable colors for scripting
  ```

#### `template`
- **Type**: String
- **Default**: `""`
- **Description**: Custom Go template for output formatting
- **Environment**: `IT2_TEMPLATE`
- **Advanced usage only**

### Retry Settings

#### `retry_attempts`
- **Type**: Integer
- **Default**: `3`
- **Description**: Number of retry attempts for failed API calls
- **Environment**: `IT2_RETRY_ATTEMPTS`
- **Examples**:
  ```yaml
  retry_attempts: 5  # Retry up to 5 times
  retry_attempts: 0  # Disable retries
  ```

#### `retry_delay`
- **Type**: Duration string
- **Default**: `1s`
- **Description**: Delay between retry attempts
- **Environment**: `IT2_RETRY_DELAY`
- **Examples**:
  ```yaml
  retry_delay: "2s"    # 2 second delay
  retry_delay: "500ms" # 500ms delay
  ```

### Logging Settings

#### `log_level`
- **Type**: String
- **Default**: `info`
- **Options**: `debug`, `info`, `warn`, `error`
- **Description**: Minimum log level to display
- **Environment**: `IT2_LOG_LEVEL`
- **Examples**:
  ```yaml
  log_level: "debug"  # Show all log messages
  log_level: "error"  # Only show errors
  ```

#### `verbose`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable verbose output
- **Environment**: `IT2_VERBOSE`
- **Examples**:
  ```yaml
  verbose: true  # Show detailed operation info
  ```

#### `debug`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Enable debug output (very verbose)
- **Environment**: `IT2_DEBUG` or `ITERM2_DEBUG`
- **Examples**:
  ```yaml
  debug: true  # Show debug information
  ```

### Progress Settings

#### `show_progress`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Show progress indicators for long operations
- **Environment**: `IT2_SHOW_PROGRESS`
- **Examples**:
  ```yaml
  show_progress: false  # Disable progress bars
  ```

### Authentication Settings

#### `cookie` and `key`
- **Type**: String
- **Default**: `""`
- **Description**: Authentication credentials (auto-managed)
- **Environment**: `ITERM2_COOKIE`, `ITERM2_KEY`
- **Note**: Usually managed automatically, don't set manually

## Environment Variables

### Complete List

| Variable | Config Key | Type | Description |
|----------|------------|------|-------------|
| `IT2_URL` | `url` | string | WebSocket URL |
| `IT2_TIMEOUT` | `timeout` | duration | Connection timeout |
| `IT2_FORMAT` | `format` | string | Output format |
| `IT2_COLOR` | `color` | boolean | Enable colors |
| `IT2_TEMPLATE` | `template` | string | Custom template |
| `IT2_RETRY_ATTEMPTS` | `retry_attempts` | int | Retry count |
| `IT2_RETRY_DELAY` | `retry_delay` | duration | Retry delay |
| `IT2_LOG_LEVEL` | `log_level` | string | Log level |
| `IT2_VERBOSE` | `verbose` | boolean | Verbose output |
| `IT2_DEBUG` | `debug` | boolean | Debug output |
| `IT2_SHOW_PROGRESS` | `show_progress` | boolean | Progress bars |
| `ITERM2_COOKIE` | `cookie` | string | Auth cookie |
| `ITERM2_KEY` | `key` | string | Auth key |
| `ITERM2_DEBUG` | `debug` | boolean | Debug (alias) |

### Setting Environment Variables

**Bash/Zsh:**
```bash
export IT2_FORMAT=json
export IT2_TIMEOUT=10s
```

**Fish:**
```fish
set -x IT2_FORMAT json
set -x IT2_TIMEOUT 10s
```

**Session-specific:**
```bash
IT2_FORMAT=json it2 session list
```

## Configuration Examples

### Development Environment

```yaml
# ~/.it2/config.yaml - Development setup
url: "ws://localhost:1912"
timeout: "30s"
format: "json"
verbose: true
debug: false
retry_attempts: 1
show_progress: false
```

### Production/Scripting

```yaml
# ~/.it2/config.yaml - Production/scripting setup
url: "ws://localhost:1912"
timeout: "5s"
format: "json"
color: false
verbose: false
debug: false
retry_attempts: 0
show_progress: false
log_level: "error"
```

### Remote iTerm2

```yaml
# ~/.it2/config.yaml - Remote iTerm2 access
url: "ws://remote-server:1912"
timeout: "15s"
format: "table"
retry_attempts: 5
retry_delay: "2s"
```

### Debug Configuration

```yaml
# ~/.it2/config.yaml - Debug setup
url: "ws://localhost:1912"
timeout: "60s"
format: "yaml"
verbose: true
debug: true
log_level: "debug"
show_progress: true
```

## Profile-Specific Configuration

Create different configuration files for different use cases:

```bash
# Development profile
cp ~/.it2/config.yaml ~/.it2/config-dev.yaml

# Production profile
cp ~/.it2/config.yaml ~/.it2/config-prod.yaml

# Use specific config
IT2_CONFIG=~/.it2/config-dev.yaml it2 session list
```

## Shell Integration

### Bash/Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# it2 configuration
export IT2_FORMAT=table
export IT2_TIMEOUT=10s

# Aliases for common configurations
alias it2-json='IT2_FORMAT=json it2'
alias it2-debug='IT2_DEBUG=1 it2'
alias it2-remote='IT2_URL=ws://remote:1912 it2'

# Function to quickly switch configs
it2_config() {
    case "$1" in
        dev)
            export IT2_CONFIG=~/.it2/config-dev.yaml
            ;;
        prod)
            export IT2_CONFIG=~/.it2/config-prod.yaml
            ;;
        default)
            unset IT2_CONFIG
            ;;
        *)
            echo "Usage: it2_config {dev|prod|default}"
            ;;
    esac
}
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
# it2 configuration
set -x IT2_FORMAT table
set -x IT2_TIMEOUT 10s

# Aliases
alias it2-json='IT2_FORMAT=json it2'
alias it2-debug='IT2_DEBUG=1 it2'
```

## Validation

### Validate Configuration

```bash
# Check if configuration is valid
it2 config show

# Test with temporary override
IT2_TIMEOUT=1s it2 session list
```

### Common Issues

1. **YAML syntax errors**: Use proper indentation and quotes
2. **Invalid duration format**: Use Go duration format (`30s`, `5m`, `1h`)
3. **Boolean values**: Use `true`/`false`, not `yes`/`no` or `1`/`0`
4. **File permissions**: Ensure config file is readable

### Debug Configuration Loading

```bash
# See which config values are being used
IT2_DEBUG=1 it2 config show
```

## Security Considerations

### Authentication Credentials
- Don't commit `cookie` or `key` values to version control
- These are auto-managed by it2
- Delete `~/.it2/` if credentials become corrupted

### Remote Connections
- Use secure networks for remote iTerm2 connections
- Consider SSH tunneling for remote access:
  ```bash
  ssh -L 1912:localhost:1912 remote-host
  it2 session list  # Uses tunneled connection
  ```

### File Permissions
```bash
# Secure configuration directory
chmod 700 ~/.it2/
chmod 600 ~/.it2/config.yaml
```

## Migration and Backup

### Backup Configuration
```bash
# Backup entire configuration
tar -czf ~/.it2-backup.tar.gz ~/.it2/

# Backup just config file
cp ~/.it2/config.yaml ~/.it2/config.yaml.backup
```

### Migration Between Machines
```bash
# Export configuration
scp ~/.it2/config.yaml user@newhost:~/.it2/

# Import and verify
ssh user@newhost 'it2 config show'
```

This configuration system provides flexible control over it2 behavior while maintaining backwards compatibility and ease of use.