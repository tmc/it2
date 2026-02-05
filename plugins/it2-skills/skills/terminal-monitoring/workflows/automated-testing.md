# Automated Testing Workflow

Using it2 for terminal-based testing and CI/CD integration.

## Scenario

Run tests in a monitored session and capture results.

## Basic Test Runner

```bash
#!/bin/bash

PROJECT_DIR="$1"
TEST_CMD="${2:-go test ./...}"

# Create test session
TEST_SID=$(it2 session split --horizontal -q)
it2 session set-badge "$TEST_SID" "$(echo $TEST_SID | cut -c1-8)\nTests"

# Run tests
it2 session send-text "$TEST_SID" "cd '$PROJECT_DIR' && $TEST_CMD"

# Wait for completion (requires shell integration)
echo "Waiting for tests to complete..."

# Poll for prompt return
while true; do
  SCREEN=$(it2 session get-screen "$TEST_SID")
  if echo "$SCREEN" | tail -1 | grep -qE '^\$|^%|^>'; then
    break
  fi
  sleep 1
done

# Capture output
OUTPUT=$(it2 session get-buffer "$TEST_SID" --lines 100)

# Check for failures
if echo "$OUTPUT" | grep -q "FAIL"; then
  echo "Tests FAILED"
  it2 notification "Tests failed in $PROJECT_DIR"
  exit 1
else
  echo "Tests PASSED"
  exit 0
fi
```

## Parallel Test Sessions

Run tests for multiple packages concurrently:

```bash
#!/bin/bash

PACKAGES=("./pkg/api" "./pkg/db" "./pkg/auth")
declare -A SESSIONS

# Create session for each package
for pkg in "${PACKAGES[@]}"; do
  SID=$(it2 session split --horizontal -q)
  SESSIONS["$pkg"]="$SID"

  NAME=$(basename "$pkg")
  it2 session set-badge "$SID" "$(echo $SID | cut -c1-8)\n$NAME"
  it2 session send-text "$SID" "go test $pkg -v"
done

# Wait for all to complete
FAILED=()
for pkg in "${PACKAGES[@]}"; do
  SID="${SESSIONS[$pkg]}"

  while true; do
    if it2 session get-screen "$SID" | tail -1 | grep -qE '^\$|^%'; then
      break
    fi
    sleep 1
  done

  if it2 session get-buffer "$SID" --lines 50 | grep -q "FAIL"; then
    FAILED+=("$pkg")
  fi
done

if [ ${#FAILED[@]} -gt 0 ]; then
  echo "Failed packages: ${FAILED[*]}"
  exit 1
fi

echo "All tests passed"
```

## Continuous Watch Mode

```bash
#!/bin/bash

PROJECT_DIR="$1"
WATCH_SID=$(it2 session split --vertical -q)

it2 session set-badge "$WATCH_SID" "$(echo $WATCH_SID | cut -c1-8)\nWatch"
it2 session send-text "$WATCH_SID" "cd '$PROJECT_DIR'"

# Start test watcher
it2 session send-text "$WATCH_SID" "go test ./... -v"

# Monitor for failures
it2 session tail -f "$WATCH_SID" | while read line; do
  if echo "$line" | grep -q "FAIL"; then
    it2 notification "Test failure detected"
    it2 attention --type fireworks "$WATCH_SID"
  fi
done
```

## Integration with CI

Export test results:

```bash
#!/bin/bash

# Run tests and capture to file
TEST_SID=$(it2 session split -q)
OUTPUT_FILE="/tmp/test-output-$(date +%s).txt"

it2 session send-text "$TEST_SID" "go test ./... -v 2>&1 | tee $OUTPUT_FILE; echo 'TEST_COMPLETE'"

# Wait for completion marker
while ! it2 session get-screen "$TEST_SID" | grep -q 'TEST_COMPLETE'; do
  sleep 1
done

# Parse results
if grep -q "PASS" "$OUTPUT_FILE"; then
  echo "::set-output name=status::success"
else
  echo "::set-output name=status::failure"
fi
```

## Best Practices

1. **Use badges** to identify test sessions visually
2. **Capture output** to files for later analysis
3. **Monitor for patterns** rather than polling exit codes when shell integration unavailable
4. **Send notifications** on failures for unattended runs
5. **Clean up sessions** after tests complete
