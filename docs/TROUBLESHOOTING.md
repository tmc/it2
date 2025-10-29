# it2 Troubleshooting Guide

Common issues and solutions when using the it2 CLI tool.

## Connection Issues

### Problem: "failed to connect" error

```
Error: failed to connect: websocket: bad handshake
```

**Solutions:**
1. **Check if iTerm2 API is enabled:**
   - iTerm2 → Preferences → General → Magic
   - Enable "Enable Python API"

2. **Verify iTerm2 is running:**
   ```bash
   # Check if iTerm2 is running
   ps aux | grep iTerm2
   ```

3. **Check WebSocket URL:**
   ```bash
   # Default URL should work for local iTerm2
   it2 --url ws://localhost:1912 session list
   ```

4. **Test with different timeout:**
   ```bash
   it2 --timeout 10s session list
   ```

### Problem: Authentication failures

```
Error: authentication failed
```

**Solutions:**
1. **Allow API access when prompted:**
   - iTerm2 will show a dialog asking for permission
   - Click "OK" to allow API access

2. **Check authentication cookies:**
   ```bash
   # Check if auth cookies exist
   ls -la ~/.it2/

   # Remove and retry if corrupted
   rm -rf ~/.it2/
   it2 session list
   ```

3. **Manual authentication:**
   ```bash
   it2 auth request
   ```

## Configuration Issues

### Problem: Config file not being used

**Check config location:**
```bash
it2 config path
```

**Verify config contents:**
```bash
it2 config show
```

**Create default config:**
```bash
it2 config init
```

### Problem: Environment variables not working

**Check variable names:**
- `IT2_URL` for WebSocket URL
- `IT2_TIMEOUT` for connection timeout
- `IT2_FORMAT` for output format
- `IT2_DEBUG` for debug output

**Test with explicit values:**
```bash
IT2_URL=ws://localhost:1912 it2 session list
```

## Plugin Issues

### Problem: Plugin not working/discovered

**Check plugin discovery:**
```bash
# Enable debug to see plugin discovery
export ITERM2_DEBUG=1
it2 session list
```

**Verify plugin in PATH:**
```bash
# Check if plugin is executable and in PATH
which it2-session-myplugin
ls -la $(which it2-session-myplugin)
```

**Test plugin directly:**
```bash
# Test with sample arguments
./it2-session-myplugin "test-session" "test-name"
echo $?  # Should be 0
```

### Problem: Plugin timeout

**Plugins have a 5-second timeout. To fix:**

1. **Optimize plugin performance:**
   - Cache expensive operations
   - Exit early for irrelevant sessions
   - Use efficient commands

2. **Test plugin timing:**
   ```bash
   time ./it2-session-myplugin "session-id" "session-name"
   ```

3. **Add timeout handling to plugin:**
   ```bash
   #!/bin/bash
   # Add timeout to external commands
   timeout 3s some-expensive-command
   ```

## Output Format Issues

### Problem: JSON output malformed

**Use proper format flag:**
```bash
it2 --format json session list
```

**Validate JSON output:**
```bash
it2 --format json session list | jq '.'
```

### Problem: Table formatting issues

**For wide tables:**
```bash
# Use text format for better readability
it2 --format text session list

# Or pipe to less
it2 session list | less -S
```

**For custom formatting:**
```bash
# Use YAML for human-readable structure
it2 --format yaml session list
```

## send-text Delivery Issues

### Problem: "Text partially delivered" warnings

```
⚠ Text partially delivered (some characters may be missing)
```

**Understanding exit codes:**
- Exit 0: Success (text delivered and confirmed)
- Exit 1: Error (connection failure, invalid arguments)
- Exit 2: Partial delivery (some text delivered, retryable)
- Exit 3: No delivery (session busy/modal, retryable)
- Exit 4: Modal detected (not safe to send)

**Solutions:**

1. **Skip verification for speed (recommended for automation):**
   ```bash
   it2 session send-text SESSION_ID --skip-confirm "command"
   ```

2. **Enable automatic retry on transient failures:**
   ```bash
   it2 session send-text SESSION_ID --retry 3 --retry-delay 2s "command"
   ```

3. **Wait for session to be ready before sending:**
   ```bash
   it2 session send-text SESSION_ID --require is-at-prompt,has-no-partial-input "command"
   ```

4. **Debug delivery confirmation:**
   ```bash
   IT2_DEBUG_DELIVERY=1 it2 session send-text SESSION_ID "test"
   ```
   This shows exactly what text is being matched and why delivery confirmation succeeds or fails.

5. **Combine strategies for reliable automation:**
   ```bash
   it2 session send-text SESSION_ID \
     --require is-at-prompt \
     --retry 3 \
     --retry-delay 2s \
     "make build"
   ```

**Note:** The "partially delivered" warning can be a false positive, especially with interactive sessions. If you're certain the text was delivered (check with `it2 session get-screen`), use `--skip-confirm` to bypass verification.

### Problem: Text not appearing in session

**Check session state:**
```bash
# Get current screen contents
it2 session get-screen SESSION_ID

# Check if session is at a prompt
it2 session send-text SESSION_ID --require is-at-prompt --verbose "test"
```

**Common causes:**
- Session has a modal dialog open (use `--require` preconditions)
- Session is running a TUI application (vim, htop, etc.)
- Session has partial input that conflicts with sent text

**Solutions:**
```bash
# Wait for session to be ready
it2 session send-text SESSION_ID --require has-no-partial-input "command"

# Check for specific session states
it2 session send-text SESSION_ID \
  --require is-at-prompt,has-no-partial-input \
  --require-timeout 30s \
  "command"
```

## Session Management Issues

### Problem: Session ID not found

```
Error: session not found
```

**Check session exists:**
```bash
it2 session list
```

**Use session completion:**
```bash
# Enable shell completion first
it2 completion bash | source

# Then use tab completion
it2 session activate <TAB>
```

**Check if using correct session ID format:**
- Session IDs are typically UUIDs or iTerm2-generated strings
- Don't confuse with session names

### Problem: Cannot create sessions

**Check profile exists:**
```bash
it2 profile list
```

**Use explicit profile:**
```bash
it2 session create --profile "Default"
```

**Check window context:**
```bash
# List windows first
it2 window list

# Create in specific window
it2 session create --window-id <window-id>
```

## Shell Integration Issues

### Problem: Shell integration features not working

```
Error: command not found - shell integration required
```

**Enable Shell Integration:**
1. Install iTerm2 Shell Integration:
   ```bash
   curl -L https://iterm2.com/shell_integration/install_shell_integration_and_utilities.sh | bash
   ```

2. Restart your shell or source the integration:
   ```bash
   source ~/.iterm2_shell_integration.bash  # or .zsh
   ```

**Verify integration is working:**
```bash
# Check if environment variable is set
echo $ITERM_SESSION_ID

# Test shell integration commands
it2 prompt list
```

## Performance Issues

### Problem: Commands are slow

**Use shorter timeouts for testing:**
```bash
it2 --timeout 1s session list
```

**Disable plugins temporarily:**
```bash
# Move plugins out of PATH temporarily
mkdir ~/plugins-backup
mv ~/.local/bin/it2-* ~/plugins-backup/
```

**Check network connectivity:**
```bash
# Test WebSocket connection directly
nc -z localhost 1912
```

## Debug Mode

### Enable comprehensive debugging:

```bash
export ITERM2_DEBUG=1
export IT2_DEBUG=1
it2 session list
```

This will show:
- Plugin discovery process
- WebSocket communication
- Authentication steps
- Timing information

### Increase verbosity:

```bash
it2 --format text session list 2>&1 | tee debug.log
```

## Shell Completion Issues

### Problem: Completion not working

**Install completion for your shell:**

**Bash:**
```bash
it2 completion bash > ~/.it2-completion.bash
echo "source ~/.it2-completion.bash" >> ~/.bashrc
source ~/.bashrc
```

**Zsh:**
```bash
it2 completion zsh > ~/.it2-completion.zsh
echo "source ~/.it2-completion.zsh" >> ~/.zshrc
source ~/.zshrc
```

**Fish:**
```bash
it2 completion fish > ~/.config/fish/completions/it2.fish
```

### Problem: Completion shows no results

**Check iTerm2 connection:**
- Completion requires active iTerm2 connection
- Verify basic commands work: `it2 session list`

**Test completion manually:**
```bash
# Should show available sessions
it2 session activate <TAB><TAB>
```

## Log Files and Debugging

### Check system logs:

**macOS Console.app:**
- Filter for "iTerm2" or "it2"
- Look for WebSocket or API related errors

**Command line logs:**
```bash
# iTerm2 logs
log show --predicate 'process == "iTerm2"' --last 1h

# System WebSocket logs
log show --predicate 'message CONTAINS "websocket"' --last 1h
```

### it2 debug output:

Save debug output for analysis:
```bash
ITERM2_DEBUG=1 it2 session list > debug-output.log 2>&1
```

## Common Error Messages

### "connection refused"
- iTerm2 not running or API disabled
- Check iTerm2 preferences for Python API

### "timeout exceeded"
- Increase timeout: `--timeout 10s`
- Check network connectivity
- Plugin might be hanging

### "invalid session ID"
- Session may have been closed
- Refresh session list: `it2 session list`
- Use tab completion for valid IDs

### "profile not found"
- Check available profiles: `it2 profile list`
- Use exact profile name with quotes if needed

### "permission denied"
- Authentication issue
- Clear auth cookies: `rm -rf ~/.it2/`
- Restart iTerm2 and retry

## Getting Help

### Built-in help:
```bash
it2 --help
it2 session --help
it2 session create --help
```

### Version information:
```bash
it2 version  # If implemented
```

### Report issues:
Include this information when reporting bugs:
- Operating system and version
- iTerm2 version
- it2 command that failed
- Full error message
- Debug output (with `ITERM2_DEBUG=1`)

### Community resources:
- Check existing issues on GitHub
- Include minimal reproduction steps
- Provide configuration and environment details