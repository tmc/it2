# Complex Layout Recipes

Patterns for creating multi-pane terminal layouts.

## Standard Layouts

### IDE-Style (Editor + Terminal + Output)

```
┌─────────────────┬────────────────┐
│                 │                │
│     Editor      │    Output      │
│                 │                │
├─────────────────┴────────────────┤
│           Terminal               │
└──────────────────────────────────┘
```

```bash
MAIN=$ITERM_SESSION_ID

# Create right pane (output)
OUTPUT=$(it2 session split --vertical -q)

# Create bottom pane from main (terminal)
TERM=$(it2 session split --horizontal -q)

# Label each
it2 session set-badge "$MAIN" "$(echo $MAIN | cut -c1-8)\nEditor"
it2 session set-badge "$OUTPUT" "$(echo $OUTPUT | cut -c1-8)\nOutput"
it2 session set-badge "$TERM" "$(echo $TERM | cut -c1-8)\nTerminal"
```

### Four Quadrants

```
┌────────────┬────────────┐
│     1      │     2      │
├────────────┼────────────┤
│     3      │     4      │
└────────────┴────────────┘
```

```bash
Q1=$ITERM_SESSION_ID
Q2=$(it2 session split --vertical -q)
Q3=$(it2 session split --horizontal -q)  # Splits Q1
it2 session focus "$Q2"
Q4=$(it2 session split --horizontal -q)  # Splits Q2
```

### Server Monitor (Many Horizontal)

```
┌──────────────────────────────────┐
│           Server 1               │
├──────────────────────────────────┤
│           Server 2               │
├──────────────────────────────────┤
│           Server 3               │
└──────────────────────────────────┘
```

```bash
SERVERS=("prod-1" "prod-2" "prod-3")
MAIN=$ITERM_SESSION_ID
SESSIONS=("$MAIN")

for ((i=1; i<${#SERVERS[@]}; i++)); do
  SID=$(it2 session split --horizontal -q)
  SESSIONS+=("$SID")
done

for ((i=0; i<${#SERVERS[@]}; i++)); do
  it2 session set-badge "${SESSIONS[$i]}" "$(echo ${SESSIONS[$i]} | cut -c1-8)\n${SERVERS[$i]}"
done
```

## Size Considerations

Before creating layouts, check available space:

```bash
# Get grid size
it2 session get-info --format json | jq .grid_size
```

Typical minimums for usable panes:
- Width: 80 columns
- Height: 24 rows

## Layout Persistence

Save current layout as arrangement:

```bash
it2 arrangement save "my-layout"

# Restore later
it2 arrangement restore "my-layout"
```
