# iTerm2 Python API vs it2 CLI: Feature Coverage Comparison

**Generated:** 2025-10-01
**Python API Version:** Latest (as of iterm2.com documentation)
**it2 CLI Version:** Current (main branch)

## Executive Summary

The `it2` CLI tool provides comprehensive coverage of core iTerm2 functionality with a focus on command-line automation and scripting. This report compares the Python API capabilities with current `it2` CLI coverage, identifies gaps, and suggests improvements.

### Coverage Overview

- ✅ **Excellent Coverage:** Session, tab, window management, text operations, shell integration
- ⚠️ **Partial Coverage:** Status bar, color management, profiles
- ❌ **Missing:** Long-running daemons, hooks system, custom RPCs, status bar component creation

---

## Detailed Feature Comparison

### 1. Core Application Control

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Get app properties | ✅ `app.async_get_variable()` | ✅ `it2 app get-property` | ✅ Full | Complete coverage |
| Set app variables | ✅ `app.async_set_variable()` | ✅ `it2 variable set` | ✅ Full | Complete coverage |
| Focus app | ✅ `app.async_activate()` | ✅ `it2 app focus` | ✅ Full | Complete coverage |
| Get all windows | ✅ `app.windows` | ✅ `it2 window list` | ✅ Full | Complete coverage |
| Get all sessions | ✅ `app.sessions` | ✅ `it2 session list` | ✅ Full | Complete coverage |

### 2. Session Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List sessions | ✅ `app.sessions` | ✅ `it2 session list` | ✅ Full | Excellent formatting options |
| Create session | ✅ `window.async_create_tab()` | ✅ `it2 tab create` | ✅ Full | Complete coverage |
| Close session | ✅ `session.async_close()` | ✅ `it2 session close` | ✅ Full | Complete coverage |
| Split pane | ✅ `session.async_split_pane()` | ✅ `it2 session split` | ✅ Full | Vertical/horizontal support |
| Get session variables | ✅ `session.async_get_variable()` | ✅ `it2 variable get session` | ✅ Full | Complete coverage |
| Set session variables | ✅ `session.async_set_variable()` | ✅ `it2 variable set session` | ✅ Full | Complete coverage |
| Activate session | ✅ `session.async_activate()` | ✅ `it2 session activate` | ✅ Full | Complete coverage |
| Restart session | ✅ `session.async_restart()` | ✅ `it2 session restart` | ✅ Full | Complete coverage |
| Get current session | ✅ Via connection | ✅ `it2 session current` | ✅ Full | Uses `$ITERM_SESSION_ID` |
| Tree view | ❌ Not applicable | ✅ `it2 session tree` | ✅ Better | CLI enhancement |

### 3. Text and Buffer Operations

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Send text | ✅ `session.async_send_text()` | ✅ `it2 session send-text` | ✅ Full | Complete coverage |
| Send keys | ✅ Via custom sequences | ✅ `it2 session send-key` | ✅ Full | Named keys support |
| Get buffer contents | ✅ `screen.contents` | ✅ `it2 text get-buffer` | ✅ Full | Complete coverage |
| Search buffer | ✅ `screen.async_search()` | ✅ `it2 text search` | ✅ Full | Complete coverage |
| Clear buffer | ✅ Custom sequence | ✅ `it2 text clear` | ✅ Full | Complete coverage |
| Get selection | ✅ `session.async_get_selection()` | ✅ `it2 selection get` | ✅ Full | Complete coverage |
| Set selection | ✅ `session.async_set_selection()` | ✅ `it2 selection set` | ✅ Full | Complete coverage |
| Get selection text | ✅ `async_get_selection_text()` | ✅ `it2 selection get-text` | ✅ Full | Complete coverage |
| Cursor position | ✅ `screen.cursor_coord` | ✅ `it2 text get-cursor` | ✅ Full | Complete coverage |
| Move cursor | ✅ Via escape sequences | ✅ `it2 text set-cursor` | ✅ Full | Complete coverage |
| Inject text at cursor | ✅ `session.async_inject()` | ✅ `it2 text inject` | ✅ Full | Complete coverage |

### 4. Window and Tab Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List windows | ✅ `app.windows` | ✅ `it2 window list` | ✅ Full | Complete coverage |
| Create window | ✅ `app.async_create_window()` | ✅ `it2 window create` | ✅ Full | Complete coverage |
| Close window | ✅ `window.async_close()` | ✅ `it2 window close` | ✅ Full | Complete coverage |
| Get window properties | ✅ `window.frame` | ✅ `it2 window get-property` | ✅ Full | Complete coverage |
| Set window frame | ✅ `window.async_set_frame()` | ✅ `it2 window set-frame` | ✅ Full | Complete coverage |
| List tabs | ✅ `window.tabs` | ✅ `it2 tab list` | ✅ Full | Complete coverage |
| Create tab | ✅ `window.async_create_tab()` | ✅ `it2 tab create` | ✅ Full | Complete coverage |
| Close tab | ✅ `tab.async_close()` | ✅ `it2 tab close` | ✅ Full | Complete coverage |
| Activate tab | ✅ `tab.async_activate()` | ✅ `it2 tab activate` | ✅ Full | Complete coverage |
| Move tab | ✅ `tab.async_move()` | ✅ `it2 tab move` | ✅ Full | Complete coverage |
| Get tab variables | ✅ `tab.async_get_variable()` | ✅ `it2 variable get tab` | ✅ Full | Complete coverage |
| Set tab color | ✅ `change.set_tab_color()` | ⚠️ Partial | ⚠️ Partial | Via variable, not direct API |

### 5. Profile Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List profiles | ✅ `app.async_list_profiles()` | ✅ `it2 profile list` | ✅ Full | Complete coverage |
| Get profile properties | ✅ Via partial profile | ✅ `it2 profile get-property` | ✅ Full | Complete coverage |
| Set profile properties | ✅ `session.async_set_profile_property()` | ✅ `it2 profile set-property` | ✅ Full | Complete coverage |
| Set session profile | ✅ `session.async_set_profile()` | ✅ `it2 session set-profile` | ✅ Full | Complete coverage |
| Create/delete profiles | ❌ Not in API | ❌ Not supported | ❌ Missing | iTerm2 limitation |
| Partial profile changes | ✅ `PartialProfile` | ✅ Via set-property | ✅ Full | Complete coverage |

### 6. Color Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List color presets | ✅ `app.async_list_color_presets()` | ✅ `it2 color list` | ✅ Full | Complete coverage |
| Get current preset | ✅ `session.async_get_variable()` | ✅ `it2 color get` | ✅ Full | Complete coverage |
| Set color preset | ✅ `session.async_set_color_preset()` | ✅ `it2 color set` | ✅ Full | Complete coverage |
| Import color preset | ❌ Not direct | ⚠️ Via file | ⚠️ Workaround | Requires manual import |
| Export color preset | ❌ Not direct | ❌ Not supported | ❌ Missing | Could be added |
| Get individual colors | ✅ Via profile properties | ✅ `it2 profile get-property` | ✅ Full | Complete coverage |

### 7. Shell Integration Features

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List prompts | ✅ Via variables | ✅ `it2 prompt list` | ✅ Full | Excellent implementation |
| Search command history | ❌ Manual parsing | ✅ `it2 prompt search` | ✅ Better | CLI enhancement |
| Get prompt info | ✅ Via variables | ✅ `it2 prompt get` | ✅ Full | Complete coverage |
| List jobs | ✅ Via variables | ✅ `it2 job list` | ✅ Full | Complete coverage |
| Monitor jobs | ✅ Via notifications | ✅ `it2 job monitor` | ✅ Full | Complete coverage |

### 8. Clipboard Operations

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Copy to clipboard | ✅ Via pasteboard | ✅ `it2 clipboard copy` | ✅ Full | Complete coverage |
| Paste from clipboard | ✅ Via pasteboard | ✅ `it2 clipboard paste` | ✅ Full | Complete coverage |
| Copy to base64 | ✅ Manual encoding | ✅ `it2 clipboard copy --base64` | ✅ Better | CLI enhancement |

### 9. Notification and Event Monitoring

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Keystroke monitoring | ✅ `async_monitor_keystroke()` | ✅ `it2 notification monitor --type keystroke` | ✅ Full | Complete coverage |
| Session updates | ✅ `async_subscribe_to_new_session_notification()` | ✅ `it2 notification monitor --type session` | ✅ Full | Complete coverage |
| Screen updates | ✅ `async_subscribe_to_screen_update()` | ✅ `it2 notification monitor --type screen` | ✅ Full | Complete coverage |
| Variable changes | ✅ `async_monitor_variable()` | ✅ `it2 variable monitor` | ✅ Full | Complete coverage |
| Tab layout changes | ✅ Notification API | ✅ `it2 notification monitor --type layout` | ✅ Full | Complete coverage |
| Profile changes | ✅ Notification API | ✅ `it2 notification monitor --type profile` | ✅ Full | Complete coverage |

### 10. Broadcast Domains

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List domains | ✅ Via API | ✅ `it2 broadcast list` | ✅ Full | Complete coverage |
| Create domain | ✅ API call | ✅ `it2 broadcast create` | ✅ Full | Complete coverage |
| Delete domain | ✅ API call | ✅ `it2 broadcast delete` | ✅ Full | Complete coverage |
| Set domain sessions | ✅ API call | ✅ `it2 broadcast set` | ✅ Full | Complete coverage |
| Send to domain | ✅ `suppress_broadcast=False` | ✅ Via domain targeting | ✅ Full | Complete coverage |

### 11. Arrangement Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List arrangements | ✅ `app.async_list_saved_arrangements()` | ✅ `it2 arrangement list` | ✅ Full | Complete coverage |
| Save arrangement | ✅ `app.async_save_arrangement()` | ✅ `it2 arrangement save` | ✅ Full | Complete coverage |
| Restore arrangement | ✅ `app.async_restore_arrangement()` | ✅ `it2 arrangement restore` | ✅ Full | Complete coverage |

### 12. tmux Integration

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| List tmux connections | ✅ `app.tmux_connections` | ✅ `it2 tmux list` | ✅ Full | Complete coverage |
| Send tmux command | ✅ `connection.async_send_command()` | ✅ `it2 tmux send-command` | ✅ Full | Complete coverage |
| Set tmux window panes | ✅ API methods | ⚠️ Limited | ⚠️ Partial | Basic support only |

### 13. Badge Management

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Set badge | ✅ Via profile property | ✅ `it2 badge set` | ✅ Full | Complete coverage |
| Get badge | ✅ Via profile property | ✅ `it2 badge get` | ✅ Full | Complete coverage |
| Clear badge | ✅ Set to empty | ✅ `it2 badge clear` | ✅ Full | Complete coverage |

### 14. Authentication

| Feature | Python API | it2 CLI | Status | Notes |
|---------|------------|---------|--------|-------|
| Request auth | ✅ Automatic | ✅ `it2 auth request` | ✅ Full | Complete coverage |
| Check auth status | ✅ Connection check | ✅ `it2 auth check` | ✅ Full | Complete coverage |
| Store credentials | ✅ Environment | ✅ Environment | ✅ Full | Complete coverage |

---

## Missing Features (Python API Capabilities Not in it2 CLI)

### 1. Long-Running Daemons ❌

**Python API:**
```python
async def main(connection):
    app = await iterm2.async_get_app(connection)
    # Script runs forever, monitoring events
    while True:
        await asyncio.sleep(1)

iterm2.run_forever(main)
```

**it2 CLI:** Not supported - CLI commands are transactional, not persistent

**Impact:** High - prevents status bar components, custom hooks, persistent monitoring
**CLI Applicability:** Low - fundamentally incompatible with CLI model
**Recommendation:** Document that daemon features require Python API

---

### 2. Status Bar Component Creation ❌

**Python API:**
```python
@iterm2.StatusBarRPC
async def my_component(knobs):
    return "Custom Status"

component = iterm2.StatusBarComponent(
    short_description="My Component",
    detailed_description="Shows custom info",
    knobs=[],
    exemplar="Example",
    update_cadence=1,
    identifier="com.example.component"
)

await component.async_register(connection, my_component)
```

**it2 CLI:** Only `it2 statusbar list` to view existing components

**Impact:** High - can't create custom status bar components
**CLI Applicability:** Low - requires persistent process
**Recommendation:** Keep view-only support, document Python API for creation

---

### 3. Custom RPC Registration ❌

**Python API:**
```python
@iterm2.RPC
async def my_function(session_id, arg1, arg2):
    # Custom function callable via key binding
    return result

await my_function.async_register(connection)
```

**it2 CLI:** Not supported - no persistent process for RPC handling

**Impact:** Medium - can't create custom key-bindable functions
**CLI Applicability:** Low - requires daemon process
**Recommendation:** Consider wrapper mechanism or document Python API requirement

---

### 4. Session Title Provider Hooks ❌

**Python API:**
```python
@iterm2.TitleProviderRPC
async def my_title_provider(session_id):
    # Dynamically generate session title
    return custom_title

await my_title_provider.async_register(connection)
```

**it2 CLI:** Can read/set titles but not provide dynamic generation

**Impact:** Medium - can't implement dynamic title generation
**CLI Applicability:** Low - requires daemon process
**Recommendation:** Document Python API for hooks

---

### 5. Custom Context Menu Items ❌

**Python API:**
```python
async def my_menu_action(session_id):
    # Handle menu click
    pass

menu_item = iterm2.MenuItemCommand(
    title="My Action",
    identifier="com.example.menu.action"
)

await menu_item.async_register(connection, my_menu_action)
```

**it2 CLI:** Not supported

**Impact:** Low - niche feature
**CLI Applicability:** Low - requires daemon process
**Recommendation:** Document Python API requirement

---

### 6. Screen Annotations ❌

**Python API:**
```python
# Add visual annotations to screen content
await session.async_add_annotation(...)
```

**it2 CLI:** Not supported

**Impact:** Low - rarely used feature
**CLI Applicability:** Medium - could be one-shot command
**Recommendation:** Consider adding if demand exists

---

### 7. Custom Escape Sequence Handling ❌

**Python API:**
```python
@iterm2.CustomControlSequence
async def handle_sequence(session_id, data):
    # Handle custom OSC sequences
    pass
```

**it2 CLI:** Not supported

**Impact:** Low - very advanced use case
**CLI Applicability:** Low - requires daemon process
**Recommendation:** Document Python API requirement

---

## Partial Coverage Features (Could Be Enhanced)

### 1. Tab Color Management ⚠️

**Current:** Can set via variable, but indirect
**Enhancement:** Add `it2 tab set-color <rgb>` direct command

### 2. Profile Color Editing ⚠️

**Current:** Must set individual color properties
**Enhancement:** Add bulk color operations, preset templates

### 3. Screen Capture ⚠️

**Current:** Basic buffer retrieval
**Enhancement:** Add image capture, HTML export options

### 4. tmux Integration ⚠️

**Current:** Basic command sending
**Enhancement:** Add tmux-specific operations (pane management, etc.)

---

## it2 CLI Advantages Over Python API

### 1. Better Command History Integration ✨

**it2 CLI:**
```bash
it2 prompt search "git commit" --exit-code 0
it2 prompt list --format json | jq '.[] | select(.exit_code != 0)'
```

More powerful filtering and formatting than manual Python parsing.

### 2. Tree Visualization ✨

**it2 CLI:**
```bash
it2 session tree
```

Visual hierarchy not directly available in Python API.

### 3. Output Format Flexibility ✨

All commands support `--format json|yaml|table|text` for easy scripting integration.

### 4. Authentication Simplicity ✨

Automatic auth handling with clear error messages and fallback mechanisms.

### 5. Session ID Normalization ✨

Handles both full and UUID-only formats transparently, plus automatic `$ITERM_SESSION_ID` fallback.

---

## Recommendations for it2 CLI Improvements

### Priority 1: High Value, Medium Effort

1. **Direct Tab Color API**
   - Add: `it2 tab set-color <tab-id> <r> <g> <b>`
   - Add: `it2 tab get-color <tab-id>`
   - Removes need for variable manipulation

2. **Session Plugins API**
   - Current: `it2 plugins` command exists but underdocumented
   - Enhancement: Better documentation and examples
   - Add plugin list/enable/disable operations

3. **Screen Capture Enhancements**
   - Add: `it2 screen capture --format [text|html|ansi]`
   - Add: `it2 screen capture --lines N` for partial capture
   - Better than current text-only buffer retrieval

4. **Enhanced Color Preset Management**
   - Add: `it2 color export <preset> <file>`
   - Add: `it2 color import <file>`
   - Add: `it2 color create <name> --from <preset>`

### Priority 2: Quality of Life

5. **Bulk Operations**
   - Add: `it2 session send-text-all <text>` for all sessions
   - Add: `it2 tab close-all --except <id>`
   - Add: `it2 session activate-by-name <pattern>`

6. **Better History Operations**
   - Add: `it2 prompt export --format csv`
   - Add: `it2 prompt stats` for command frequency analysis
   - Enhancement: Time range filtering

7. **Window Management**
   - Add: `it2 window tile` for automatic layout
   - Add: `it2 window minimize/maximize`
   - Better window positioning shortcuts

8. **Profile Templates**
   - Add: `it2 profile copy <source> <dest>`
   - Add: `it2 profile diff <profile1> <profile2>`
   - Quick profile cloning

### Priority 3: Advanced Features

9. **Session Recording**
   - Add: `it2 session record <output.cast>` for asciinema format
   - Add: `it2 session replay <input.cast>`
   - Better than manual buffer capture

10. **Macro System**
    - Add: `it2 macro define <name> <commands...>`
    - Add: `it2 macro run <name>`
    - Store common operation sequences

11. **Enhanced Notification Filters**
    - Add: `it2 notification monitor --filter 'expr'`
    - Add: `it2 notification log --output <file>`
    - Better event processing

12. **Arrangement Diffing**
    - Add: `it2 arrangement diff <name1> <name2>`
    - Add: `it2 arrangement merge`
    - Help manage complex setups

### Not Recommended

These Python API features are **not suitable** for CLI implementation:

- **Long-running daemons** - Fundamentally incompatible with CLI model
- **Custom status bar components** - Requires persistent process
- **RPC registration** - Needs persistent daemon
- **Title provider hooks** - Needs persistent daemon
- **Custom escape sequences** - Requires daemon
- **Context menu items** - Requires daemon

For these features, users should use the Python API directly.

---

## Implementation Strategy

### Phase 1: Direct API Gaps (2-3 weeks)
1. Direct tab color commands
2. Screen capture enhancements
3. Color preset import/export
4. Plugin management improvements

### Phase 2: Quality of Life (2-3 weeks)
5. Bulk operation commands
6. Enhanced history operations
7. Window management shortcuts
8. Profile template commands

### Phase 3: Advanced Features (3-4 weeks)
9. Session recording/replay
10. Macro system
11. Enhanced notification filtering
12. Arrangement diffing tools

---

## Documentation Improvements

### Add New Sections

1. **"When to Use Python API vs it2 CLI"**
   - Clear guidance on use case selection
   - Feature compatibility matrix
   - Migration patterns

2. **"Creating Status Bar Components"**
   - Document Python API requirement
   - Provide example scripts
   - Link to Python API docs

3. **"Advanced Automation Patterns"**
   - Combining CLI and Python API
   - Event-driven workflows
   - Best practices

4. **"API Coverage Matrix"**
   - Include this comparison document
   - Keep updated with releases
   - Add CLI version to API mapping

---

## Testing Gaps

Based on Python API examples, recommend adding integration tests for:

1. **Targeted input** - Broadcasting with exclusions
2. **Tab color persistence** - Verify across restarts
3. **Profile property edge cases** - Invalid values, type mismatches
4. **Notification filtering** - Complex filter expressions
5. **Long-running operations** - Timeout handling, cleanup
6. **Error recovery** - Connection loss, auth failure, invalid IDs

---

## Conclusion

The `it2` CLI provides **excellent coverage** of core iTerm2 functionality suitable for command-line automation. Key strengths:

✅ **Complete:** Session/tab/window management, text operations, shell integration
✅ **Enhanced:** Better command history, tree views, output formatting
✅ **Robust:** Authentication, error handling, session ID normalization

Key gaps are **fundamental architecture differences**:

❌ No persistent daemon features (status bar components, hooks, RPCs)
⚠️ Some advanced features need better APIs (colors, screen capture, profiles)

**Recommended Priority:**
1. Direct color/tab APIs (Phase 1)
2. Screen capture enhancements (Phase 1)
3. Bulk operations (Phase 2)
4. Document Python API for daemon features

The CLI should **not attempt** to replicate daemon-based features - instead, improve documentation showing when Python API is the right choice.

---

## Appendix: Python API Example Coverage

| Example Script | Functionality | it2 CLI Equivalent | Status |
|----------------|---------------|-------------------|--------|
| cls.py | Clear screen | `it2 text clear` | ✅ Full |
| settabcolor.py | Set tab color | `it2 variable set` (indirect) | ⚠️ Partial |
| function_key_tabs.py | Key binding handler | Not applicable (daemon) | ❌ N/A |
| jsonpretty.py | Status bar component | Not applicable (daemon) | ❌ N/A |
| targeted_input.py | Broadcast with exclusions | `it2 session send-text` + targeting | ✅ Full |
| theme.py | Cycle color presets | `it2 color set` | ✅ Full |
| movetab.py | Move tab between windows | `it2 tab move` | ✅ Full |
| georges_title.py | Dynamic titles | Not applicable (daemon) | ❌ N/A |
| zoom.py | Zoom session | `it2 session zoom` | ✅ Full |
| broadcast.py | Manage broadcast domains | `it2 broadcast` commands | ✅ Full |
| tmux_basic.py | tmux integration | `it2 tmux` commands | ✅ Full |

**Coverage Summary:**
- Transactional operations: **90% covered**
- Daemon features: **0% covered** (by design)
- Overall: **~70% functional coverage** (appropriate for CLI tool)
