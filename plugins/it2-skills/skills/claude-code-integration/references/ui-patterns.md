# Claude Code UI Patterns

Reference for parsing Claude Code terminal UI output.

## Message Structure

Claude Code displays messages in bordered boxes:

```
╭─ Claude ─────────────────────────────────────────────╮
│ Message content here                                 │
╰──────────────────────────────────────────────────────╯
```

## Tool Use Display

When Claude uses tools:

```
╭─ Tool: Read ─────────────────────────────────────────╮
│ Reading: /path/to/file.go                            │
╰──────────────────────────────────────────────────────╯
```

## Status Indicators

| Character | Meaning |
|-----------|---------|
| `●` | Actively working (animated) |
| `○` | Idle/waiting |
| `?` | Needs input |
| `✓` | Success |
| `✗` | Error |
| `⋯` | Loading/processing |

## Permission Prompts

Claude requests permission with patterns like:

```
Allow Claude to run this command?
  $ some-command --flag

  [Y]es  [N]o  [A]lways allow
```

Detection:
```bash
if it2 session get-screen "$SID" | grep -q 'Allow Claude to'; then
  echo "Permission prompt detected"
fi
```

## Progress Display

Long-running operations show progress:

```
Processing: ████████░░░░░░░░ 50%
```

## Code Blocks

Code is displayed with syntax highlighting context:

```
╭─ Edit: file.go ──────────────────────────────────────╮
│ func main() {                                        │
│     fmt.Println("Hello")                             │
│ }                                                    │
╰──────────────────────────────────────────────────────╯
```

## Extracting Content

Get just the message content:

```bash
# Extract text between box borders
it2 session get-screen "$SID" | \
  sed -n '/╭─/,/╰─/p' | \
  grep -v '^[╭╰│]'
```

## Error Patterns

Errors typically show:

```
╭─ Error ──────────────────────────────────────────────╮
│ Command failed with exit code 1                      │
│ stderr: error message here                           │
╰──────────────────────────────────────────────────────╯
```

Detection:
```bash
if it2 session get-screen "$SID" | grep -q '╭─ Error'; then
  echo "Error detected in Claude output"
fi
```
