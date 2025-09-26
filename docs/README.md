# it2 Documentation

Welcome to the it2 documentation! it2 is a comprehensive command-line interface for controlling iTerm2, providing programmatic access to iTerm2's powerful features.

## Quick Links

- [Quick Start Guide](QUICKSTART.md) - Get up and running in minutes
- [Configuration](CONFIGURATION.md) - Configure it2 for your environment
- [Examples](EXAMPLES.md) - Practical examples and use cases
- [Automation Cookbook](AUTOMATION_COOKBOOK.md) - Recipes for common automation tasks

## What is it2?

it2 is a command-line tool that provides complete control over iTerm2 through its API. With it2, you can:

- **Manage Sessions**: Create, control, and monitor terminal sessions
- **Automate Tasks**: Send commands, capture output, and respond to events
- **Control Windows & Tabs**: Programmatically manage your iTerm2 workspace
- **Handle Text & Buffers**: Search, replace, and manipulate terminal content
- **Configure Profiles**: Manage and apply iTerm2 profiles dynamically
- **Monitor Events**: Subscribe to and react to iTerm2 notifications

## Getting Started

1. **Install it2**:
   ```bash
   go install github.com/tmc/it2/cmd/it2@latest
   ```

2. **Enable iTerm2 API**:
   - Open iTerm2 Preferences
   - Go to General → Magic → Enable Python API
   - Restart iTerm2

3. **Verify Connection**:
   ```bash
   it2 auth check
   ```

## Command Structure

it2 commands follow a hierarchical structure:

```
it2 <resource> <action> [options]
```

For example:
- `it2 session list` - List all sessions
- `it2 tab create` - Create a new tab
- `it2 text send "echo hello"` - Send text to current session

## Documentation Structure

### User Guides
- [Quick Start](QUICKSTART.md) - Essential commands to get started
- [Examples](EXAMPLES.md) - Real-world usage examples
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions

### Reference
- [Configuration](CONFIGURATION.md) - Environment variables and settings
- [Plugin Development](PLUGIN_DEVELOPMENT.md) - Extending it2 functionality
- [Automation Cookbook](AUTOMATION_COOKBOOK.md) - Advanced automation recipes

### Command Reference
Browse the complete command reference in the sidebar, organized by resource type (session, tab, window, etc.)

## Key Features

### Session Management
Control iTerm2 sessions with commands like:
- `it2 session create` - Create new sessions
- `it2 session send-text` - Send commands
- `it2 session monitor` - Watch for events

### Window & Tab Control
Manage your workspace:
- `it2 window create` - Open new windows
- `it2 tab split` - Split panes
- `it2 arrangement save` - Save layouts

### Text Operations
Powerful text manipulation:
- `it2 text search` - Search buffer content
- `it2 text replace` - Find and replace
- `it2 selection copy` - Copy selected text

### Profile Management
Dynamic profile control:
- `it2 profile list` - View available profiles
- `it2 profile set` - Apply profiles to sessions
- `it2 profile export` - Export profile settings

## Contributing

it2 is open source! Visit the [GitHub repository](https://github.com/tmc/it2) to:
- Report issues
- Submit pull requests
- View source code
- Star the project

## Support

- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [GitHub Issues](https://github.com/tmc/it2/issues)
- [Documentation](DOCUMENTATION.md)

---

*Built with ❤️ for iTerm2 power users*