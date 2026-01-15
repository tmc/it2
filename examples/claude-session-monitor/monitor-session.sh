#!/bin/bash

# Enhanced self-monitoring script for single Claude session with modal detection
#
# This script monitors a Claude Code session and automatically intervenes when:
#   - Session has incomplete todos and is idle at prompt
#   - Session shows inefficient sleep polling patterns
#   - Session has vague monitoring todos without milestones
#   - Claude approval modals appear (Do you want to proceed?)
#   - Choice modals with numbered options appear
#
# It will NOT intervene when session shows active indicators:
#   - Screen contents actively changing
#   - User appears to be typing
#   - Spinners and work symbols (✳ ✽ ✶ ✢)
#   - 'esc to interrupt' or 'ctrl+t to hide todos'
#   - 'Generating', 'Processing', 'Working', etc.

SESSION_ID="${1:-$ITERM_SESSION_ID}"
MONITOR_INTERVAL=5
MIN_ACTION_INTERVAL=30
MODAL_COOLDOWN=15  # Prevent rapid modal re-detection
STABILITY_CHECK_ENABLED=false  # Use it2's built-in stability detection
PARTIAL_TEXT_TIMEOUT=180  # After 3 minutes of partial text, send it
CHANGE_WINDOW_SIZE=3  # Track last N changes to detect sustained activity
CHANGE_THRESHOLD=100  # Characters that must change to count as "changing"
AUTO_CONTINUE_INTERVAL=300  # Send 'continue' every 5 minutes if todos exist
LAST_CONTENT=""
LAST_ACTION_TIME=0
LAST_MODAL_TIME=0
LAST_MODAL_HASH=""
LAST_STABILITY_TIME=0
LAST_PARTIAL_TEXT=""
PARTIAL_TEXT_FIRST_SEEN=0
LAST_AUTO_CONTINUE_TIME=$(date +%s)  # Initialize to now so first auto-continue waits full interval
RECENT_CHANGES=()  # Array to track recent change sizes

# Create short session identifier for logging
SESSION_SHORT=$(echo "$SESSION_ID" | cut -c1-8)

# Color codes for better visibility
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Handle help flag
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Claude Session Self-Monitor Enhanced (Single Session)"
    echo "Usage: $0 [SESSION_ID] [--help]"
    echo ""
    echo "Monitors a Claude Code session and automatically intervenes when:"
    echo "  - Session has incomplete todos (☐) and is idle at prompt (>)"
    echo "  - Session shows inefficient sleep polling patterns"
    echo "  - Session has vague monitoring todos without milestones"
    echo "  - Claude approval modals appear (Do you want to proceed?)"
    echo "  - Choice modals with numbered options appear"
    echo ""
    echo "WILL NOT intervene when session shows active indicators:"
    echo "  - Screen contents actively changing (using it2 stability detection)"
    echo "  - User appears to be typing (multiple heuristics)"
    echo "  - ✳ ✽ ✶ ✢ (spinners and work symbols)"
    echo "  - 'esc to interrupt' or 'ctrl+t to hide todos'"
    echo "  - 'Generating', 'Processing', 'Working', 'Puttering', 'Moseying', etc."
    echo ""
    echo "Smart Detection Features:"
    echo "  - Screen stability detection using it2 --wait-for-stability"
    echo "  - Multi-pattern typing detection (handles wrapped prompts)"
    echo "  - Content change tracking (won't interrupt if anything changed)"
    echo "  - 60s idle requirement before intervening (conservative)"
    echo "  - Partial text timeout: sends Enter after ${PARTIAL_TEXT_TIMEOUT}s of unchanged text"
    echo ""
    echo "Modal Detection Features:"
    echo "  - Detects 'Do you want to proceed?' prompts"
    echo "  - Handles choice dialogs with '❯ 1. Yes' patterns"
    echo "  - Recognizes box drawing characters for modal borders"
    echo ""
    echo "Sets session title to 'watching-[target]' during monitoring"
    echo "Check interval: ${MONITOR_INTERVAL}s, min action: ${MIN_ACTION_INTERVAL}s"
    exit 0
fi

# Set monitoring badge
MONITOR_SESSION_ID="${ITERM_SESSION_ID}"
if [ -n "$MONITOR_SESSION_ID" ] && [ -n "$SESSION_ID" ]; then
    MONITORED_SHORT=$(echo "$SESSION_ID" | cut -c1-8)
    # todo: set badge
    # it2 session set-title "$MONITOR_SESSION_ID" "watching-$MONITORED_SHORT"
fi

echo -e "${CYAN}Starting enhanced single-session monitor${NC}"
echo -e "${BLUE}Target Session:${NC} $SESSION_ID (${YELLOW}$SESSION_SHORT${NC})"
echo -e "${BLUE}Check interval:${NC} ${MONITOR_INTERVAL}s, ${BLUE}Min action:${NC} ${MIN_ACTION_INTERVAL}s"
echo -e "${GREEN}Modal detection: ENABLED${NC}"
echo -e "${GREEN}Screen stability detection: ENABLED${NC} (using it2 --wait-for-stability)"
echo -e "${GREEN}Multi-pattern typing detection: ENABLED${NC}"
echo -e "${GREEN}Partial text timeout: ${PARTIAL_TEXT_TIMEOUT}s${NC} (will send Enter after this timeout)"
echo "---"

# Logging function with session context
log_action() {
    local action_type="$1"
    local message="$2"
    local color="${3:-$NC}"
    echo -e "[$(date '+%H:%M:%S')] ${color}[$SESSION_SHORT]${NC} ${color}$action_type:${NC} $message"
}

# Cleanup function
cleanup() {
    log_action "SHUTDOWN" "Enhanced monitor shutting down" "$RED"
    if [ -n "$MONITOR_SESSION_ID" ]; then
        sleep 0.1
        # todo: set badge
        # it2 session set-title "$MONITOR_SESSION_ID" ""
    fi
    exit 0
}

trap cleanup EXIT SIGINT SIGTERM

# Function to detect Claude approval modals
detect_approval_modal() {
    local content="$1"

    # Pattern 1: "Do you want to proceed?" style prompts
    if echo "$content" | grep -qiE "Do you want to proceed\?|Proceed\?|Continue\?"; then
        return 0
    fi

    # Pattern 2: Choice modals with "❯ 1. Yes" patterns
    if echo "$content" | grep -qE "❯[[:space:]]*[0-9]+\.[[:space:]]*(Yes|Continue|Proceed)"; then
        return 0
    fi

    # Pattern 3: Simple Yes/No prompts
    if echo "$content" | grep -qE "\[Y/n\]|\[y/N\]|\(y/n\)"; then
        return 0
    fi

    return 1
}

# NOTE: Claude has 3 high-level modes shown in the status bar:
#   1. Normal mode - regular conversation/coding
#   2. Accept edits mode - Claude is trying to make edits (shows "accept edits on")
#   3. Plan mode - planning/designing
# These are MODE indicators, not MODAL states. The presence of "accept edits on"
# doesn't mean Claude is waiting for user input - it just indicates the current mode.
# Previously this script incorrectly treated "accept edits" as a modal trigger.


# Function to detect choice modals with numbered options
detect_choice_modal() {
    local content="$1"

    # Pattern 1: Multiple numbered options in modal boxes
    # Look for numbered options inside box characters (│) with or without arrow
    # Format: "│ ❯ 1. Yes" or "│   2. No"
    local option_count
    option_count=$(echo "$content" | grep -E "[│][[:space:]]*(❯[[:space:]]*)?[0-9]+\.[[:space:]]+(Yes|No)" | wc -l | tr -d ' ')
    if [ "$option_count" -ge 2 ]; then
        return 0
    fi

    # Pattern 2: Look for "Do you want to proceed?" with numbered options
    # This catches Claude Code approval dialogs which always have at least 2 options
    if echo "$content" | grep -q "Do you want to proceed?" && \
       echo "$content" | grep -qE "❯.*[0-9]+\.[[:space:]]*Yes"; then
        # Found the prompt and a Yes option, now count total numbered options
        local num_options
        num_options=$(echo "$content" | grep -E "[0-9]+\.[[:space:]]+(Yes|No)" | wc -l | tr -d ' ')
        if [ "$num_options" -ge 2 ]; then
            return 0
        fi
    fi

    # Pattern 3: Choice indicators with arrows (need multiple)
    local arrow_count
    arrow_count=$(echo "$content" | grep -E "→[[:space:]]*[0-9]+\.|▶[[:space:]]*[0-9]+\." | wc -l | tr -d ' ')
    if [ "$arrow_count" -ge 2 ]; then
        return 0
    fi

    # Pattern 4: Actual modal box with corners AND numbered options nearby
    # Only match if we see top corners AND at least one numbered option
    if echo "$content" | grep -qE "╭.*╮|┌.*┐" && echo "$content" | grep -qE "❯.*[0-9]+\."; then
        return 0
    fi

    return 1
}

# Function to extract partial text from prompt line
# Returns the text after the prompt, or empty string if none
get_partial_text() {
    local content="$1"

    # Find line starting with > and extract text after it
    # Use [[:space:]] for portability (BSD sed doesn't support \s)
    local partial=$(echo "$content" | awk '/^>/ {print}' | tail -1 | sed 's/^>[[:space:]]*//')

    # Trim all whitespace (leading and trailing)
    partial=$(echo "$partial" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

    # Filter out known Claude UI messages (not actual user input)
    # These are shown in grey/dimmed text in the actual terminal
    if echo "$partial" | grep -qE "^Press up to|^Press .* to edit|^queued messages|^accept edits|^shift\+tab"; then
        return 1
    fi

    # Only return if there's actual non-whitespace content
    if [ -n "$partial" ]; then
        echo "$partial"
        return 0
    fi

    return 1
}

# Function to check if screen is actively changing (not stable)
# Returns 0 (success) if screen IS changing, 1 (failure) if stable
# This matches bash convention where 0=true for use in if statements
# Sets SCREEN_CHANGE_INFO with details about the change
is_screen_changing() {
    if [ "$STABILITY_CHECK_ENABLED" != "true" ]; then
        SCREEN_CHANGE_INFO=""
        return 1  # Assume stable (not changing) if check disabled
    fi

    # Get current screen and compare with last check
    local current_screen=$(it2 session get-screen "$SESSION_ID" 2>/dev/null)
    local change_chars=0

    if [ -n "$LAST_SCREEN_FOR_DIFF" ]; then
        # Calculate approximate change size (number of differing characters)
        change_chars=$(diff <(echo "$LAST_SCREEN_FOR_DIFF") <(echo "$current_screen") | wc -c | tr -d ' ')
    fi

    LAST_SCREEN_FOR_DIFF="$current_screen"

    # Add to recent changes window (keep last N)
    RECENT_CHANGES+=($change_chars)
    if [ ${#RECENT_CHANGES[@]} -gt $CHANGE_WINDOW_SIZE ]; then
        RECENT_CHANGES=("${RECENT_CHANGES[@]:1}")  # Remove oldest
    fi

    # Calculate average change over window
    local total=0
    local count=${#RECENT_CHANGES[@]}
    for size in "${RECENT_CHANGES[@]}"; do
        total=$((total + size))
    done
    local avg_change=$((count > 0 ? total / count : 0))

    # Screen is "changing" only if average change over window exceeds threshold
    # This filters out minor fluctuations like token counters
    if [ $avg_change -gt $CHANGE_THRESHOLD ]; then
        SCREEN_CHANGE_INFO="~${avg_change}ch/avg (${RECENT_CHANGES[*]})"
        return 0  # Screen IS changing (return success for if statement)
    fi

    SCREEN_CHANGE_INFO="~${change_chars}ch (stable)"
    return 1  # Screen is stable (return failure for if statement)
}

# Function to handle modal interactions
handle_modal() {
    local content="$1"

    # Generate hash of modal content to track if it's the same modal
    local modal_hash
    modal_hash=$(echo "$content" | grep -E "Do you want|Proceed|❯|Yes|No" | md5)

    # If this is the same modal we just tried to dismiss, don't retry immediately
    if [ "$modal_hash" = "$LAST_MODAL_HASH" ]; then
        local time_since_modal=$((CURRENT_TIME - LAST_MODAL_TIME))
        if [ $time_since_modal -lt $MODAL_COOLDOWN ]; then
            log_action "MODAL" "Same modal within cooldown (${time_since_modal}s) - waiting" "$YELLOW"
            return 1
        fi
    fi

    # If there's a numbered option with arrow, send "1" then Enter
    # Otherwise just send Enter (default option)
    if echo "$content" | grep -q "❯"; then
        log_action "MODAL" "Sending '1' + Enter" "$GREEN"
        it2 session send-text "$SESSION_ID" --send-cr=false "1"
        sleep 0.5
        it2 session send-key "$SESSION_ID" "Return"
    else
        log_action "MODAL" "Sending Enter" "$GREEN"
        it2 session send-key "$SESSION_ID" "Return"
    fi

    LAST_MODAL_HASH="$modal_hash"
    LAST_MODAL_TIME=$CURRENT_TIME
    sleep 2

    # Verify
    local post_content
    post_content=$(it2 session get-screen "$SESSION_ID" 2>/dev/null)
    if [ $? -eq 0 ]; then
        if echo "$post_content" | grep -q "Do you want to proceed"; then
            log_action "MODAL" "Still present" "$YELLOW"
        else
            log_action "MODAL" "Dismissed" "$GREEN"
        fi
    fi

    return 0
}

# Main monitoring loop
while true; do
    # Get current content
    CURRENT_CONTENT=$(it2 session get-screen "$SESSION_ID" 2>/dev/null)

    if [ $? -ne 0 ]; then
        log_action "ERROR" "Failed to get session content" "$RED"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Check timing
    CURRENT_TIME=$(date +%s)
    TIME_SINCE_LAST=$((CURRENT_TIME - LAST_ACTION_TIME))

    # Priority -1: Check if screen is actively changing using diff-based detection
    # This is the most reliable way to detect active output/generation
    if is_screen_changing; then
        # Screen is actively changing - skip all checks
        log_action "CHANGING" "Screen actively changing ${SCREEN_CHANGE_INFO} - skipping all interventions" "$BLUE"
        sleep $MONITOR_INTERVAL
        continue
    else
        # Screen is stable - log occasionally (every 30s)
        if [ $((CURRENT_TIME % 30)) -eq 0 ]; then
            log_action "WATCHING" "Monitoring stable screen ${SCREEN_CHANGE_INFO}" "$CYAN"
        fi
    fi

    # Priority 0: Check for active work indicators FIRST - don't interrupt active sessions
    # "esc to interrupt" is the most reliable indicator that Claude Code is actively working
    # Check last 15 lines to catch spinners above todo lists
    RECENT_LINES=$(echo "$CURRENT_CONTENT" | tail -15)
    # Note: "esc to interrupt" pattern must NOT match menu options like "(esc)" in choice dialogs
    # Use word boundary \b or require it to be a complete phrase
    # Note: ⏺ removed - it's a UI marker shown in command previews, not an activity indicator
    if echo "$RECENT_LINES" | grep -qE "\besc to interrupt\b|ctrl\+t to hide|Sketching|Generating|Working|✳|✽|✶|✢"; then
        log_action "SKIP" "Session has active work - skipping all interventions" "$BLUE"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Extract recent content once for all checks (last 20 lines)
    RECENT_CONTENT=$(echo "$CURRENT_CONTENT" | tail -20)

    # Priority 1: Check for approval modals BEFORE checking partial text
    # Modals can appear over partial text, so they take precedence
    # IMPORTANT: Only check recent content, not scrollback history
    if detect_approval_modal "$RECENT_CONTENT"; then
        if handle_modal "$RECENT_CONTENT"; then
            LAST_ACTION_TIME=$CURRENT_TIME
        fi
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Check for partial user input - don't interrupt if user is typing
    # Look for prompt with non-empty text after it (user typing a command)
    # Check the last 5 lines to catch wrapped prompts and multi-line input
    LAST_LINES=$(echo "$CURRENT_CONTENT" | tail -5)

    # Extract partial text now (after modal checks)
    PARTIAL_TEXT=$(get_partial_text "$CURRENT_CONTENT")

    if [ -n "$PARTIAL_TEXT" ]; then
        # Normalize for comparison (removes all leading/trailing whitespace)
        # This prevents session focus changes from resetting the timer
        PARTIAL_TEXT_NORMALIZED=$(echo "$PARTIAL_TEXT" | xargs)
        LAST_PARTIAL_NORMALIZED=$(echo "$LAST_PARTIAL_TEXT" | xargs)

        # We have partial text - track it
        # Compare normalized versions to ignore whitespace from focus/selection changes
        if [ "$PARTIAL_TEXT_NORMALIZED" = "$LAST_PARTIAL_NORMALIZED" ] && [ -n "$LAST_PARTIAL_NORMALIZED" ]; then
            # Same partial text as before (ignoring whitespace)
            if [ $PARTIAL_TEXT_FIRST_SEEN -eq 0 ]; then
                # First time seeing this partial text
                PARTIAL_TEXT_FIRST_SEEN=$CURRENT_TIME
                log_action "PARTIAL" "Detected partial text: '$PARTIAL_TEXT' - tracking" "$CYAN"
            else
                # We've been tracking this partial text
                TIME_PARTIAL_WAITING=$((CURRENT_TIME - PARTIAL_TEXT_FIRST_SEEN))

                if [ $TIME_PARTIAL_WAITING -ge $PARTIAL_TEXT_TIMEOUT ]; then
                    # Partial text has been sitting for too long - send it
                    log_action "SEND" "Partial text waited ${TIME_PARTIAL_WAITING}s (>= ${PARTIAL_TEXT_TIMEOUT}s) - sending Enter" "$YELLOW"
                    it2 session send-key "$SESSION_ID" "Return"
                    LAST_ACTION_TIME=$CURRENT_TIME
                    PARTIAL_TEXT_FIRST_SEEN=0
                    LAST_PARTIAL_TEXT=""
                    sleep $MONITOR_INTERVAL
                    continue
                else
                    # Still waiting, but log progress occasionally
                    TIME_REMAINING=$((PARTIAL_TEXT_TIMEOUT - TIME_PARTIAL_WAITING))
                    if [ $((TIME_PARTIAL_WAITING % 30)) -eq 0 ]; then
                        log_action "WAITING" "Partial text '$PARTIAL_TEXT' - will press Enter in ${TIME_REMAINING}s" "$CYAN"
                    fi
                fi
            fi
        else
            # Different partial text than before - reset tracking
            PARTIAL_TEXT_FIRST_SEEN=$CURRENT_TIME
            LAST_PARTIAL_TEXT="$PARTIAL_TEXT"
            log_action "PARTIAL" "New partial text detected: '$PARTIAL_TEXT' - tracking" "$CYAN"
        fi

        # Skip all other interventions while partial text is present
        sleep $MONITOR_INTERVAL
        continue
    else
        # No partial text - reset tracking
        if [ -n "$LAST_PARTIAL_TEXT" ]; then
            log_action "CLEARED" "Partial text cleared (was: '$LAST_PARTIAL_TEXT')" "$GREEN"
        fi
        PARTIAL_TEXT_FIRST_SEEN=0
        LAST_PARTIAL_TEXT=""
    fi

    # Original typing detection patterns (kept as fallback)
    # Extract any partial text for logging
    FALLBACK_PARTIAL=$(get_partial_text "$CURRENT_CONTENT")
    TYPING_CONTEXT=""
    if [ -n "$FALLBACK_PARTIAL" ]; then
        # Truncate to first 40 chars for logging
        TYPING_CONTEXT=" ('$(echo "$FALLBACK_PARTIAL" | cut -c1-40)')"
    fi

    # Pattern 1: Basic prompt patterns with text (original detection)
    if echo "$LAST_LINES" | grep -qE "^>[[:space:]]+[^[:space:]]" || echo "$LAST_LINES" | grep -qE "\$[[:space:]]+[^[:space:]]"; then
        log_action "TYPING" "User appears to be typing${TYPING_CONTEXT} - skipping all interventions" "$CYAN"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Pattern 2: Prompt anywhere in line with text (handles wrapped prompts)
    if echo "$LAST_LINES" | grep -qE ">[[:space:]]+[a-zA-Z0-9]"; then
        log_action "TYPING" "Detected text after prompt${TYPING_CONTEXT} - skipping all interventions" "$CYAN"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Pattern 3: Very last line has non-whitespace content after any prompt separator
    # This catches cases where user typed text but prompt isn't at line start
    VERY_LAST_LINE=$(echo "$CURRENT_CONTENT" | tail -1)
    if echo "$VERY_LAST_LINE" | grep -qE ">[[:space:]]*[^[:space:]]+|>[[:space:]]+[^[:space:]]"; then
        log_action "TYPING" "Found text on prompt line${TYPING_CONTEXT} - skipping all interventions" "$CYAN"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Pattern 4: Check for cursor position indicators showing active input
    # Claude shows "Puttering", "Moseying", etc when processing user input
    if echo "$LAST_LINES" | grep -qE "Puttering|Moseying|Sketching" && echo "$LAST_LINES" | grep -q ">"; then
        log_action "TYPING" "Claude processing recent user input${TYPING_CONTEXT} - skipping all interventions" "$CYAN"
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Auto-continue: If session has active todos, send 'continue' every 5 minutes
    # This keeps sessions progressing without manual intervention
    TIME_SINCE_AUTO_CONTINUE=$((CURRENT_TIME - LAST_AUTO_CONTINUE_TIME))
    if echo "$CURRENT_CONTENT" | grep -q "☐" && \
       echo "$CURRENT_CONTENT" | grep -qE "^>[[:space:]]*$|────────────" && \
       [ $TIME_SINCE_AUTO_CONTINUE -ge $AUTO_CONTINUE_INTERVAL ]; then
        # Check if session appears idle (no active work indicators)
        # Use RECENT_CONTENT (tail -20) and check for ALL active work patterns
        if ! echo "$RECENT_CONTENT" | grep -qE "esc to interrupt|ctrl\+t to hide|Sketching|Generating|Working|Thinking|Waiting|Running|✳|✽|✶|✢|∴"; then
            log_action "AUTO_CONTINUE" "Active todos detected, sending 'continue' (${TIME_SINCE_AUTO_CONTINUE}s since last)" "$GREEN"
            it2 session send-text "$SESSION_ID" --send-cr=false "continue"
            it2 session send-key "$SESSION_ID" enter
            LAST_AUTO_CONTINUE_TIME=$CURRENT_TIME
            LAST_ACTION_TIME=$CURRENT_TIME
            sleep $MONITOR_INTERVAL
            continue
        fi
    fi

    # Skip if too soon since last action (for non-modal actions)
    if [ $TIME_SINCE_LAST -lt $MIN_ACTION_INTERVAL ]; then
        sleep $MONITOR_INTERVAL
        continue
    fi

    # Check for inefficient polling
    if echo "$CURRENT_CONTENT" | grep -qE "sleep [0-9]+.*sleep [0-9]+" && \
       echo "$CURRENT_CONTENT" | grep -q "Monitor.*progress"; then
        log_action "POLLING" "Detected inefficient polling - suggesting better approach" "$YELLOW"
        it2 session send-text "$SESSION_ID" --send-cr=false "Instead of manual sleep polling, let me set up systematic milestone-based monitoring."
        it2 session send-key "$SESSION_ID" enter
        LAST_ACTION_TIME=$CURRENT_TIME

    # Check for vague todos
    elif echo "$CURRENT_CONTENT" | grep -q "☐ Monitor.*progress" && \
         ! echo "$CURRENT_CONTENT" | grep -qE "☐.*Phase|☐.*completion|☐.*milestone" && \
         echo "$CURRENT_CONTENT" | grep -qE "^>[[:space:]]*$"; then
        log_action "TODOS" "Detected vague monitoring todos - suggesting milestones" "$CYAN"
        it2 session send-text "$SESSION_ID" --send-cr=false "Let me break down monitoring into specific milestones: Phase 1, Phase 2, Phase 3 completion."
        it2 session send-key "$SESSION_ID" enter
        LAST_ACTION_TIME=$CURRENT_TIME

    # Check for API/tool failure errors - these need immediate recovery
    elif echo "$CURRENT_CONTENT" | grep -qE "API Error:|tool_use.*without.*tool_result|catastrophic.*failure"; then
        if echo "$CURRENT_CONTENT" | grep -qE "^>[[:space:]]*$"; then
            log_action "TOOL_FAILURE" "Detected API/tool failure - sending continue to recover" "$RED"
            it2 session send-text "$SESSION_ID" --send-cr=false "continue"
            it2 session send-key "$SESSION_ID" enter
            LAST_ACTION_TIME=$CURRENT_TIME
        fi

    # Check for stuck session with todos (with safety checks)
    elif echo "$CURRENT_CONTENT" | grep -q "☐" && \
         echo "$CURRENT_CONTENT" | grep -qE "^>[[:space:]]*$|────────────"; then

        # Additional safety: Check if content has changed recently (even slightly)
        # If content is different from last check, something is happening
        if [ "$CURRENT_CONTENT" != "$LAST_CONTENT" ]; then
            log_action "ACTIVE" "Content changed since last check - session is active" "$BLUE"
            # Continue to next iteration
        # Safety check: Don't send continue if recent errors or corruption (but API errors are OK)
        # Look for actual error conditions, not just the word "error" in conversation
        elif echo "$RECENT_CONTENT" | grep -qE "^Error:|^ERROR:|^Failed:|^FAILED:|^Fatal:|^FATAL:|Traceback|Exception|panic:|runtime error|segmentation fault|core dumped|command not found|permission denied|No such file|cannot find|unable to|compilation failed"; then
            # Unless it's an API error which we want to recover from
            if echo "$RECENT_CONTENT" | grep -qE "API Error:|tool_use.*without.*tool_result"; then
                log_action "CONTINUE" "API error detected with todos - sending continue to recover" "$YELLOW"
                it2 session send-text "$SESSION_ID" --send-cr=false "continue"
                it2 session send-key "$SESSION_ID" enter
                LAST_ACTION_TIME=$CURRENT_TIME
            else
                log_action "SAFETY" "Session has actual errors - NOT sending continue (safety)" "$RED"
            fi
        # Safety check: Don't send continue too frequently
        elif [ $TIME_SINCE_LAST -lt 30 ]; then
            log_action "WAIT" "Too soon since last action - waiting" "$YELLOW"
        # Safety check: Require idle for longer before intervening (60s instead of 30s)
        elif [ $TIME_SINCE_LAST -lt 60 ]; then
            log_action "WAIT" "Not idle long enough (${TIME_SINCE_LAST}s < 60s) - waiting" "$YELLOW"
        else
            log_action "CONTINUE" "Detected stuck session with todos (idle ${TIME_SINCE_LAST}s) - sending continue" "$GREEN"
            it2 session send-text "$SESSION_ID" --send-cr=false "continue"
            it2 session send-key "$SESSION_ID" enter
            LAST_ACTION_TIME=$CURRENT_TIME
        fi

    # Check for stuck on same content (but not if user is typing)
    elif [ "$CURRENT_CONTENT" = "$LAST_CONTENT" ] && [ $TIME_SINCE_LAST -gt 60 ]; then
        # Double-check: Don't send Enter if there's partial input OR an active prompt
        # Check for any of:
        # 1. Text after a prompt (> text or $ text)
        # 2. Empty prompt waiting for user input (> or $ at end of screen)
        # 3. Interactive prompt patterns (────────────, choices, etc.)
        # NOTE: Use LAST_LINES (tail -5) not RECENT_CONTENT to avoid matching old commands in history
        if echo "$LAST_LINES" | grep -qE "^>[[:space:]]+[^[:space:]]" || echo "$LAST_LINES" | grep -qE "\$[[:space:]]+[^[:space:]]"; then
            # Extract partial text for context
            STUCK_PARTIAL=$(get_partial_text "$CURRENT_CONTENT")
            if [ -n "$STUCK_PARTIAL" ]; then
                STUCK_CONTEXT=" ('$(echo "$STUCK_PARTIAL" | cut -c1-40)')"
            else
                STUCK_CONTEXT=""
            fi
            log_action "TYPING" "Content unchanged but user may be typing${STUCK_CONTEXT} - skipping" "$CYAN"
        # Check for empty prompt waiting for user (active input session)
        # Check CURRENT_CONTENT for todos since they may have scrolled out of LAST_LINES
        elif echo "$LAST_LINES" | grep -qE "^>[[:space:]]*$|────────────" && ! echo "$CURRENT_CONTENT" | grep -qE "☐"; then
            log_action "WAITING" "Active input prompt detected (no todos) - not sending Enter" "$CYAN"
        else
            log_action "STUCK" "Content unchanged for 60s - sending Enter" "$YELLOW"
            it2 session send-key "$SESSION_ID" enter
            LAST_ACTION_TIME=$CURRENT_TIME
        fi
    fi

    LAST_CONTENT="$CURRENT_CONTENT"
    sleep $MONITOR_INTERVAL
done
