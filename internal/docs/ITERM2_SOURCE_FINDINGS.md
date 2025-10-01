# iTerm2 Source Code Findings

**From Session**: 30E3 (iTerm2 source code analysis)
**To Session**: it2 CLI analysis session
**Date**: 2025-10-01

This document contains comprehensive findings from analyzing the iTerm2 source code to identify undocumented features, complete API catalogs, and implementation details that can enhance the it2 CLI tool.

---

## 1. Complete RPC Method Catalog

### All Available RPC Methods (from api.proto)

The iTerm2 WebSocket API exposes **33 distinct request types**. Here's the complete catalog:

#### Session Management
- `GetBufferRequest` - Fetch terminal buffer contents with style information
- `GetPromptRequest` - Get shell prompt metadata (requires shell integration)
- `ListPromptsRequest` - Enumerate all prompts in a session
- `GetProfilePropertyRequest` - Read profile properties
- `SetProfilePropertyRequest` - Modify profile properties
- `SendTextRequest` - Send text to session (with broadcast suppression option)
- `InjectRequest` - Inject bytes as terminal output
- `RestartSessionRequest` - Restart a session
- `CloseRequest` - Close sessions/tabs/windows

#### Tab & Window Operations
- `ListSessionsRequest` - Get all windows/tabs/sessions
- `CreateTabRequest` - Create new tab with custom properties
- `SplitPaneRequest` - Split a pane
- `SetTabLayoutRequest` - Modify split pane layout
- `ReorderTabsRequest` - Reorder tabs in windows
- `FocusRequest` - Get complete focus state snapshot
- `ActivateRequest` - Activate window/tab/session

#### Variables & Properties
- `VariableRequest` - Get/set variables in session/tab/window/app scope
- `GetPropertyRequest` - Read properties (grid_size, buried, frame, fullscreen)
- `SetPropertyRequest` - Write properties

#### Advanced Features
- `NotificationRequest` - Subscribe/unsubscribe from events
- `TransactionRequest` - Freeze/unfreeze app main loop
- `RegisterToolRequest` - Register custom toolbelt tools
- `MenuItemRequest` - Query/invoke menu items
- `SavedArrangementRequest` - Save/restore/list window arrangements
- `GetBroadcastDomainsRequest` - Get broadcast domains
- `SetBroadcastDomainsRequest` - Set broadcast domains
- `TmuxRequest` - tmux integration commands
- `PreferencesRequest` - Get/set preferences
- `ColorPresetRequest` - List/get color presets
- `SelectionRequest` - Get/set text selection
- `StatusBarComponentRequest` - Open popovers from status bar components
- `InvokeFunctionRequest` - Call built-in functions (new in recent versions)

### Undocumented/Advanced Features

#### 1. **InvokeFunctionRequest** (Recently Added)
Location: `proto/api.proto:104-158`

Allows calling built-in methods on session/tab/window/app contexts:

**Available Methods**:
- `window.set_title(title: String)`
- `session.set_name(name: String)`
- `session.run_tmux_command(command: String) throws`
- `session.set_status_bar_component_unread_count(identifier: String, count: Int)`
- `session.stop_coprocess() -> Bool`
- `session.get_coprocess() -> String?`
- `session.run_coprocess(commandLine: String, mute: Bool) -> Bool`
- `tab.set_title(title: String)`
- `tab.select_pane_in_direction(direction: String) throws -> String` - directions: 'left', 'right', 'above', 'below'

**Implementation**: This provides a more direct API for common operations that previously required variable manipulation or multiple API calls.

#### 2. **Keystroke Filtering** (Not Just Monitoring)
Location: `proto/api.proto:922-926`

The `KeystrokeFilterRequest` can prevent keystrokes from being processed by iTerm2:
- Pattern matching on modifiers, keycodes, characters
- Blocks matched keystrokes entirely
- Script receives notification and can handle as desired

**Use Case**: Implementing custom keybindings or modal editing modes in scripts.

#### 3. **Transaction API** (Performance Critical)
Location: `proto/api.proto:1304-1318`

Freezes the app's main loop during bulk operations:
- `begin=true` - Freezes time
- `begin=false` - Resumes
- Keep transactions SHORT

**Performance Insight**: Use for atomic multi-operation updates.

#### 4. **Advanced Notification Types**

Beyond basic notifications, these exist:
- `KEYSTROKE_FILTER` (14) - Does not send notification, just filters
- `NOTIFY_ON_PROFILE_CHANGE` (13) - Monitor profile changes
- `NOTIFY_ON_BROADCAST_CHANGE` (11) - Monitor broadcast domain changes
- `PromptMonitorRequest` with modes: PROMPT, COMMAND_START, COMMAND_END

#### 5. **SubSelection API**
Location: `proto/api.proto:237-245`

Supports multiple selection modes:
- CHARACTER, WORD, LINE, SMART, BOX, WHOLE_LINE
- Multiple sub-selections (rectangular selection)
- `connected` field for multi-region selections

---

## 2. Complete Variable Catalog

### Application Scope Variables (Read-Write: R/RW)

```
iTermVariableKeyApplicationPID               - RW - NSNumber - Process ID of iTerm2
iTermVariableKeyApplicationLocalhostName     - RW - NSString - Local hostname
iTermVariableKeyApplicationEffectiveTheme    - RW - NSString - Current theme ("light"/"dark")
```

### Window Scope Variables

```
iTermVariableKeyWindowTitleOverrideFormat    - RW - NSString - Format string for title
iTermVariableKeyWindowTitleOverride          - RW - NSString - Computed title override
iTermVariableKeyWindowCurrentTab             - R  - iTermVariables - Current tab object
iTermVariableKeyWindowID                     - RW - NSString - Window identifier
iTermVariableKeyWindowFrame                  - R  - NSArray[4] - [x, y, width, height]
iTermVariableKeyWindowStyle                  - R  - NSString - Window style
iTermVariableKeyWindowNumber                 - R  - NSNumber - Window number
iTermVariableKeyWindowIsHotkeyWindow         - R  - NSNumber(BOOL) - Is hotkey window
```

### Tab Scope Variables

```
iTermVariableKeyTabTitleOverride             - RW - NSString - Manual title override
iTermVariableKeyTabTitleOverrideFormat       - RW - NSString - Title format string
iTermVariableKeyTabCurrentSession            - R  - iTermVariables - Active session
iTermVariableKeyTabID                        - RW - NSString - Tab identifier
iTermVariableKeyTabTitle                     - RW - NSString - Computed tab title
iTermVariableKeyTabTmuxWindow                - RW - NSNumber - tmux window ID
iTermVariableKeyTabTmuxWindowTitle           - R  - NSString - tmux window title
iTermVariableKeyTabTmuxWindowName            - R  - NSString - tmux window name
iTermVariableKeyTabWindow                    - R  - iTermVariables - Parent window
```

### Session Scope Variables (60+ variables!)

#### Identity & Metadata
```
iTermVariableKeySessionID                    - RW - NSString - Unique session ID
iTermVariableKeySessionName                  - RW - NSString - Computed session title
iTermVariableKeySessionProfileName           - R  - NSString - Current profile name
iTermVariableKeySessionCreationTimeString    - RW - NSString - Creation timestamp
iTermVariableKeySessionAutoLogID             - RW - NSString - Auto-log file ID
iTermVariableKeySessionPresentationName      - RW - NSString - Displayed name
```

#### Location & Process Info
```
iTermVariableKeySessionHostname              - RW - NSString - Current hostname
iTermVariableKeySessionUsername              - RW - NSString - Current username
iTermVariableKeySessionPath                  - RW - NSString - Working directory
iTermVariableKeySessionHomeDirectory         - RW - NSString - $HOME (local or remote)
iTermVariableKeySessionTTY                   - RW - NSString - TTY device path
iTermVariableKeySessionTermID                - RW - NSString - Terminal ID
```

#### Job & Process Management
```
iTermVariableKeySessionJob                   - RW - NSString - Job name (unmodified)
iTermVariableKeySessionProcessTitle          - RW - NSString - Process title (argv[0])
iTermVariableKeySessionCommandLine           - RW - NSString - Foreground job + args
iTermVariableKeySessionLastCommand           - RW - NSString - Last executed command
iTermVariableKeySessionJobPid                - RW - NSNumber - Foreground job PID
iTermVariableKeySessionChildPid              - RW - NSNumber - Child process PID
iTermVariableKeySessionEffectiveSessionRootPid - RW - NSNumber - Root PID (or ssh)
```

#### Display & Appearance
```
iTermVariableKeySessionColumns               - RW - NSNumber - Terminal width
iTermVariableKeySessionRows                  - RW - NSNumber - Terminal height
iTermVariableKeySessionBadge                 - RW - NSString - Badge text (evaluated)
iTermVariableKeySessionIconName              - RW - NSString - Icon name (from escape)
iTermVariableKeySessionWindowName            - RW - NSString - Window name (from escape)
iTermVariableKeySessionAutoNameFormat        - RW - NSString - Auto-name format
iTermVariableKeySessionAutoName              - RW - NSString - Evaluated auto-name
iTermVariableKeySessionTriggerName           - RW - NSString - Trigger name
```

#### Selection & Mouse
```
iTermVariableKeySessionSelection             - RW - NSString - Selected text
iTermVariableKeySessionSelectionLength       - RW - NSNumber - Selection length
iTermVariableKeySessionMouseInfo             - RW - NSArray[7] - [x, y, button, count, modifiers, sideEffects, state]
iTermVariableKeySessionMouseReportingMode    - RW - NSNumber - Mouse mode enum
```

#### Terminal State
```
iTermVariableKeySessionShowingAlternateScreen - RW - NSNumber(BOOL) - In alt screen?
iTermVariableKeySessionApplicationKeypad      - RW - NSNumber(BOOL) - App keypad mode?
iTermVariableKeySessionBellCount              - RW - NSNumber - Times bell rang
iTermVariableKeySessionLogFilename            - RW - NSString - Log file path
```

#### tmux Integration
```
iTermVariableKeySessionTmuxRole              - RW - NSString - "", "gateway", "client"
iTermVariableKeySessionTmuxClientName        - RW - NSString - tmux session name
iTermVariableKeySessionTmuxPaneTitle         - RW - NSString - #{pane_title}
iTermVariableKeySessionTmuxWindowPane        - RW - NSNumber - #{pane_id}
iTermVariableKeySessionTmuxWindowPaneIndex   - RW - NSString - #{pane_index} (tmux 3.2+)
iTermVariableKeySessionTmuxStatusLeft        - RW - NSString - Left status (integration)
iTermVariableKeySessionTmuxStatusRight       - RW - NSString - Right status (integration)
```

#### SSH & Shell Integration
```
iTermVariableKeySSHIntegrationLevel          - RW - NSNumber - 0=none, 1=basic, 2=framer
iTermVariableKeyShell                        - RW - NSString - Basename of $SHELL
iTermVariableKeyUname                        - RW - NSString - Output of uname -a
```

#### Relationships
```
iTermVariableKeySessionTab                   - R  - iTermVariables - Parent tab
iTermVariableKeySessionParent                - R  - iTermVariables - Creating session
```

### Hidden/Power User Variables

**Note**: All user-set variables must be prefixed with `user.`

Example: `session.user.my_custom_flag`

---

## 3. Color Preset Format & Implementation

### File Format Details
Location: `sources/iTermColorPresets.h/m`

#### Structure
```
iTermColorPreset = Dictionary {
  "Color Key Name": iTermColorDictionary,
  ...
}

iTermColorDictionary = Dictionary {
  "Red Component": Number (0.0-1.0),
  "Green Component": Number (0.0-1.0),
  "Blue Component": Number (0.0-1.0),
  "Alpha Component": Number (0.0-1.0),
  "Color Space": String ("sRGB", "Calibrated", etc.)
}
```

#### All Color Keys (28 total)

**ANSI Colors (16)**:
- `Ansi 0 Color` through `Ansi 15 Color`

**Special Colors**:
- `Foreground Color`
- `Background Color`
- `Bold Color`
- `Link Color`
- `Match Background Color` - Search match highlight
- `Selection Color`
- `Selected Text Color`
- `Cursor Color`
- `Cursor Text Color`
- `Tab Color`
- `Underline Color`
- `Cursor Guide Color`
- `Badge Color`

#### Light/Dark Mode Support

Modern presets can have mode-specific variants:
- Keys can be prefixed with `Light ` or `Dark `
- Example: `Light Background Color`, `Dark Background Color`
- Function `iTermColorPresetHasModes()` checks for multi-mode presets
- Function `iTermColorPresetGet(preset, baseKey, dark)` retrieves correct variant

#### Import/Export Process

**Import**:
1. `importColorPresetFromFile:` - Read plist file
2. Parse as NSDictionary
3. Validate structure
4. Check for duplicates (perceptual color distance)
5. Store in `NSUserDefaults` under key `Custom Color Presets`
6. Post `kRebuildColorPresetsMenuNotification`

**Export**:
1. Extract color values from profile
2. Build dictionary with iTermColorDictionary entries
3. Call `iterm_writePresetToFileWithName:`
4. Writes standard plist format

**Storage**:
- Built-in presets: `ColorPresets.plist` in bundle
- Custom presets: `NSUserDefaults["Custom Color Presets"]`

### Programmatic Creation

```objective-c
// Create preset dictionary
NSDictionary *preset = @{
    @"Background Color": @{
        @"Red Component": @0.0,
        @"Green Component": @0.0,
        @"Blue Component": @0.0,
        @"Color Space": @"sRGB"
    },
    // ... more colors
};

// Add to iTerm2
[iTermColorPresets addColorPreset:@"My Theme" withColors:preset];
```

**Recommendation for it2 CLI**:
- Implement `it2 color-preset import <file>`
- Implement `it2 color-preset export <name> <file>`
- Implement `it2 color-preset create` with JSON input
- Support light/dark mode variants

---

## 4. Complete Profile Property List

### Profile Keys by Category

#### Basic Info
```
KEY_GUID                   - String - Unique identifier
KEY_NAME                   - String - Profile name
KEY_DESCRIPTION            - String - (Deprecated)
KEY_TAGS                   - Array - Tag strings
KEY_SHORTCUT               - String - Keyboard shortcut
KEY_ORIGINAL_GUID          - String - Pre-divorce GUID (not saved)
```

#### Command & Initial State
```
KEY_CUSTOM_COMMAND         - String - Use custom command
KEY_COMMAND_LINE           - String - Command to run
KEY_INITIAL_TEXT           - String - Text to send (swifty string)
KEY_INITIAL_URL            - String - URL to open
KEY_CUSTOM_DIRECTORY       - String - "Yes"/"No"/"Recycle"/"Advanced"
KEY_WORKING_DIRECTORY      - String - Working dir path
KEY_ICON                   - Number - iTermProfileIcon enum
KEY_ICON_PATH              - String - Custom icon path
```

#### Window & Display Settings
```
KEY_ROWS                   - Number - Initial rows
KEY_COLUMNS                - Number - Initial columns
KEY_WIDTH                  - Number - Width in points
KEY_HEIGHT                 - Number - Height in points
KEY_WINDOW_TYPE            - Number - Window type enum
KEY_SCREEN                 - Number - Screen number
KEY_SPACE                  - Number - macOS Space setting
KEY_USE_CUSTOM_WINDOW_TITLE - Bool - Use custom title
KEY_CUSTOM_WINDOW_TITLE    - String - Custom window title
KEY_USE_CUSTOM_TAB_TITLE   - Bool - Use custom tab title
KEY_CUSTOM_TAB_TITLE       - String - Custom tab title
KEY_TITLE_COMPONENTS       - Number - Title component flags
KEY_TITLE_FUNC             - Array - (display name, unique id) tuple
```

#### All Color Keys (Already documented above)

Plus related:
```
KEY_USE_BOLD_COLOR         - Bool - Use bold color
KEY_BRIGHTEN_BOLD_TEXT     - Bool - Brighten bold (3.3.7+)
KEY_SMART_CURSOR_COLOR     - Bool - Smart cursor color
KEY_MINIMUM_CONTRAST       - Number - Min contrast (0.0-1.0)
KEY_FAINT_TEXT_ALPHA       - Number - Faint text alpha
KEY_USE_TAB_COLOR          - Bool - Use tab color
KEY_USE_SELECTED_TEXT_COLOR - Bool - Use selected text color
KEY_USE_UNDERLINE_COLOR    - Bool - Use underline color
KEY_CURSOR_BOOST           - Number - Cursor boost amount
KEY_USE_CURSOR_GUIDE       - Bool - Show cursor guide
KEY_USE_SEPARATE_COLORS_FOR_LIGHT_AND_DARK_MODE - Bool
KEY_TRANSPARENCY_AFFECTS_ONLY_DEFAULT_BACKGROUND_COLOR - Bool
```

#### Font Settings
```
KEY_NORMAL_FONT            - String - Normal font
KEY_NON_ASCII_FONT         - String - Non-ASCII font
KEY_FONT_CONFIG            - String - Special font config
KEY_HORIZONTAL_SPACING     - Number - Horizontal spacing
KEY_VERTICAL_SPACING       - Number - Vertical spacing
KEY_USE_BOLD_FONT          - Bool - Use bold font
KEY_THIN_STROKES           - Number - Thin stroke setting
KEY_USE_ITALIC_FONT        - Bool - Use italic font
KEY_ASCII_LIGATURES        - Bool - ASCII ligatures
KEY_NON_ASCII_LIGATURES    - Bool - Non-ASCII ligatures
KEY_USE_NONASCII_FONT      - Bool - Use non-ASCII font
KEY_ASCII_ANTI_ALIASED     - Bool - ASCII anti-aliasing
KEY_NONASCII_ANTI_ALIASED  - Bool - Non-ASCII anti-aliasing
KEY_POWERLINE              - Bool - Draw powerline glyphs
```

#### Cursor Settings
```
KEY_BLINKING_CURSOR        - Bool - Cursor blinks
KEY_CURSOR_SHADOW          - Bool - Cursor has shadow
KEY_CURSOR_HIDDEN_WITHOUT_FOCUS - Bool - Hide when unfocused
KEY_CURSOR_TYPE            - Number - Cursor type enum
```

#### Transparency & Background
```
KEY_TRANSPARENCY           - Number - Transparency (0.0-1.0)
KEY_INITIAL_USE_TRANSPARENCY - Bool - Start with transparency
KEY_BLEND                  - Number - Blend amount
KEY_BLUR                   - Bool - Background blur
KEY_BLUR_RADIUS            - Number - Blur radius
KEY_BACKGROUND_IMAGE_LOCATION - String - Background image path
KEY_BACKGROUND_IMAGE_MODE  - Number - Mode enum (tiled, etc.)
```

#### Terminal Behavior
```
KEY_DISABLE_BOLD           - Bool - (Deprecated)
KEY_ANIMATE_MOVEMENT       - Bool - Animate cursor movement
KEY_BLINK_ALLOWED          - Bool - Allow blinking text
KEY_DISABLE_WINDOW_RESIZING - Bool - Disable window resize
KEY_DISABLE_UNFOCUSED_WINDOW_RESIZING - Bool
KEY_PREVENT_TAB            - Bool - Prevent opening in tab
KEY_HIDE_AFTER_OPENING     - Bool - Hide after opening
KEY_SESSION_END_ACTION     - Number - Action on session end
KEY_OPEN_TOOLBELT          - Bool - Open toolbelt
```

#### Text Encoding & Unicode
```
KEY_AMBIGUOUS_DOUBLE_WIDTH - Bool - Ambiguous chars double-width
KEY_UNICODE_NORMALIZATION  - Number - Unicode normalization mode
KEY_UNICODE_VERSION        - Number - Unicode version
```

#### Bell & Alerts
```
KEY_SILENCE_BELL           - Bool - Silence bell
KEY_VISUAL_BELL            - Bool - Visual bell
KEY_FLASHING_BELL          - Bool - Flashing bell
KEY_SEND_BELL_ALERT        - Bool - Send bell notification
KEY_SEND_IDLE_ALERT        - Bool - Send idle notification
KEY_SEND_NEW_OUTPUT_ALERT  - Bool - Send output notification
KEY_SEND_SESSION_ENDED_ALERT - Bool - Send ended notification
KEY_BOOKMARK_USER_NOTIFICATIONS - Bool - (Deprecated)
```

#### Mouse & Scrolling
```
KEY_XTERM_MOUSE_REPORTING  - Bool - Mouse reporting
KEY_XTERM_MOUSE_REPORTING_ALLOW_MOUSE_WHEEL - Bool
KEY_XTERM_MOUSE_REPORTING_ALLOW_CLICKS_AND_DRAGS - Bool
KEY_ALLOW_ALTERNATE_MOUSE_SCROLL - Bool
KEY_RESTRICT_MOUSE_REPORTING_TO_ALTERNATE_SCREEN_MODE - Bool
KEY_SCROLLBACK_WITH_STATUS_BAR - Bool
KEY_SCROLLBACK_IN_ALTERNATE_SCREEN - Bool
KEY_DRAG_TO_SCROLL_IN_ALTERNATE_SCREEN_MODE_DISABLED - Bool
```

#### Terminal Compatibility
```
KEY_DISABLE_SMCUP_RMCUP    - Bool - Disable smcup/rmcup
KEY_ALLOW_TITLE_REPORTING  - Bool - Allow title reporting
KEY_ALLOW_PASTE_BRACKETING - Bool - Bracketed paste
KEY_ALLOW_TITLE_SETTING    - Bool - Allow title setting
KEY_DISABLE_PRINTING       - Bool - Disable printing
```

#### Badge Settings
```
KEY_BADGE_FORMAT           - String - Badge text format
KEY_BADGE_TOP_MARGIN       - Number - Top margin
KEY_BADGE_RIGHT_MARGIN     - Number - Right margin
KEY_BADGE_MAX_WIDTH        - Number - Max width
KEY_BADGE_MAX_HEIGHT       - Number - Max height
KEY_BADGE_FONT             - String - Badge font
```

#### Advanced Working Directory Settings
```
KEY_AWDS_WIN_OPTION        - Number - Window option
KEY_AWDS_WIN_DIRECTORY     - String - Window directory
KEY_AWDS_TAB_OPTION        - Number - Tab option
KEY_AWDS_TAB_DIRECTORY     - String - Tab directory
KEY_AWDS_PANE_OPTION       - Number - Pane option
KEY_AWDS_PANE_DIRECTORY    - String - Pane directory
```

#### Browser Mode (Experimental)
```
KEY_BROWSER_ZOOM           - Number - Zoom level (100 = 100%)
KEY_BROWSER_DEV_NULL       - Bool - Dev null mode
KEY_BROWSER_EXTENSIONS_ROOT - String - Extensions root path
KEY_BROWSER_EXTENSION_ACTIVE_IDS - Array - Active extension IDs
```

#### Miscellaneous
```
KEY_INSTANT_REPLAY         - Bool - Instant replay enabled
KEY_PREVENT_APS            - Bool - Prevent auto profile switching
KEY_SUBTITLE               - String - Session subtitle
KEY_SSH_CONFIG             - Dictionary - SSH configuration
```

### Profile Property Types

Most properties support JSON values in the API:
- Strings: `"value"`
- Numbers: `123` or `3.14`
- Booleans: `true` / `false`
- Arrays: `["item1", "item2"]`
- Dictionaries: `{"key": "value"}`

---

## 5. Performance Optimization Opportunities

### Transaction API
**Location**: `proto/api.proto:1304-1318`

**Use Case**: Batch operations
```python
# Freeze app during bulk updates
await connection.async_transaction(begin=True)
for session in sessions:
    await session.async_set_variable("user.flag", "value")
await connection.async_transaction(begin=False)
```

### WindowedCoordRange for Buffer Access
**Location**: `proto/api.proto:223-226`, `GetBufferResponse:1181`

Instead of requesting entire buffer, use `WindowedCoordRange`:
- Specify exact coord range: `start=(x,y)` to `end=(x,y)`
- Specify column window: `columns=(location, length)`
- Returns only requested data

**Example**: Get just columns 40-80 of lines 100-200

### Metadata-Only Queries

Some properties don't require full data transfer:
- `grid_size` - Just dimensions, not content
- `number_of_lines` - Counts without content
- `buried` - Boolean flag
- `frame` - Window geometry only

### Style Information Optional
```protobuf
optional bool include_styles = 3;  // In GetBufferRequest
```

Set `include_styles=false` if you only need text content.

### Rate Limiting Considerations

**Not enforced by iTerm2**, but recommended:
- Limit variable change notifications to reasonable frequency
- Use transactions for > 10 operations
- Batch profile property changes
- Cache frequently-read values (grid_size, etc.)

---

## 6. Buffer & Screen Capture APIs

### GetBufferRequest Capabilities
**Location**: `PTYSession.m:18461-18533`, `proto/api.proto:1144-1154`

#### Request Options

1. **LineRange modes**:
   - `screen_contents_only=true` - Just visible screen
   - `trailing_lines=N` - Last N lines (includes scrollback)
   - `windowed_coord_range` - Precise coordinate range

2. **Style Information**:
   - `include_styles=true` - Full styling (colors, bold, underline, etc.)
   - CellStyle includes: fg/bg colors, bold, italic, underline, strikethrough, invisible, inverse, image placeholders, URLs, block IDs

#### Response Format

**Text Encoding**:
- UTF-8 encoded strings
- Combining characters preserved
- `code_points_per_cell` provides mapping

**Metadata**:
- Cursor position (absolute coordinates)
- Line continuation (hard EOL vs soft EOL)
- Number of lines above screen
- Windowed coord range returned

#### Advanced Features

**CellStyle Attributes** (proto/api.proto:1383-1416):
```
- Foreground: Standard (0-255), Alternate, RGB, or placement-based
- Background: Standard (0-255), Alternate, RGB, or placement-based
- Text attributes: bold, faint, italic, blink, underline, strikethrough
- Special: invisible, inverse, guarded
- Image placeholder type: NONE, ITERM2, KITTY
- External attributes: underline color, block ID, URL
- Run-length encoding via `repeats` field
```

**URL Extraction**:
CellStyle can include:
```protobuf
message URL {
  optional string url = 1;
  optional string identifier = 2;  // URL identifier for tracking
}
```

**Image Placeholders**:
```protobuf
enum ImagePlaceholderType {
  NONE = 0;
  ITERM2 = 1;
  KITTY = 2;
}
```

### Format Options We Found

**Current**: Text only
**Available but undocumented**:
- Full style information (colors, formatting)
- URL extraction
- Image placeholder detection
- Block ID tracking

**Missing formats** (would require iTerm2 enhancement):
- HTML rendering
- ANSI escape sequence reconstruction
- Metadata-only (line count, cursor pos, no content)

---

## 7. Tab Color Implementation

### Direct API Available: YES!

**Location**: Profile property `KEY_TAB_COLOR` + `KEY_USE_TAB_COLOR`

### Methods to Set Tab Color

#### Method 1: Profile Property (Recommended)
```python
# Set tab color via profile property
await session.async_set_profile_property("Tab Color", {
    "Red Component": 1.0,
    "Green Component": 0.5,
    "Blue Component": 0.0,
    "Color Space": "sRGB"
})
await session.async_set_profile_property("Use Tab Color", True)
```

#### Method 2: Variable Manipulation (Indirect)
```python
# Less reliable, more complex
tab = await session.async_get_tab()
# No direct tab.tabColor variable exposed!
```

### Implementation Details

**Profile Keys**:
- `KEY_TAB_COLOR` - iTermColorDictionary with RGBA + color space
- `KEY_USE_TAB_COLOR` - Boolean to enable/disable

**Color Space Support**:
- "sRGB" (recommended)
- "Calibrated"
- System color spaces

**Persistence**:
- Stored in session's profile copy
- Does not modify base profile
- Survives until session ends

### Recommendations for it2 CLI

Implement direct commands:
```bash
it2 tab set-color <tab-id> --rgb "255,128,0"
it2 tab set-color <tab-id> --hex "#FF8000"
it2 tab clear-color <tab-id>
it2 tab get-color <tab-id>
```

Much simpler than current variable-based approach!

---

## 8. Plugin & Extension System

### Custom RPC Registration
**Location**: `proto/api.proto:774-857`

Scripts can register custom RPCs with roles:

#### Roles

1. **GENERIC** - General-purpose RPC
2. **SESSION_TITLE** - Custom title provider
   - Attributes: `display_name`, `unique_identifier`
   - Replaces built-in title computation

3. **STATUS_BAR_COMPONENT** - Custom status bar widget
   - Attributes: `short_description`, `detailed_description`, `exemplar`, `update_cadence`
   - Knob types: Checkbox, String, PositiveFloatingPoint, Color
   - Icon support with multiple scales
   - Format: PLAIN_TEXT or HTML

4. **CONTEXT_MENU** - Custom context menu item
   - Attributes: `display_name`, `unique_identifier`

### Status Bar Component Details

**Knob Configuration**:
```protobuf
message Knob {
  optional string name = 1;
  enum Type {
    Checkbox = 1;
    String = 2;
    PositiveFloatingPoint = 3;
    Color = 4;
  }
  optional Type type = 2;
  optional string placeholder = 3;
  optional string json_default_value = 4;
  optional string key = 5;
}
```

**Icon Format**:
```protobuf
message Icon {
  optional bytes data = 1;      // PNG data
  optional float scale = 2;     // 1.0, 2.0, 3.0 for retina
}
```

### Toolbelt Tool Registration
**Location**: `proto/api.proto:748-770`

Register custom web-view based tools:
```protobuf
message RegisterToolRequest {
  optional string name = 1;                    // Display name
  optional string identifier = 2;              // Unique ID (use bundle ID)
  optional ToolType tool_type = 3;             // WEB_VIEW_TOOL
  optional bool reveal_if_already_registered = 5;
  optional string URL = 4;                     // Initial URL
}
```

**Lifecycle**:
1. First registration - Automatically added to visible tools
2. Subsequent - Only shown if `reveal_if_already_registered=true`
3. Web view has full WebKit capabilities

### Security Model

**Not explicitly documented, but inferred**:
- Toolbelt tools run in sandboxed WebView
- RPC registration requires active WebSocket connection
- No file system access beyond what WebSocket provides
- Status bar components are isolated

### Recommendations for it2 CLI

**NOT SUITABLE FOR CLI** - These require persistent daemon:
- Session title providers (need real-time updates)
- Status bar components (persistent UI)
- Context menu items (GUI integration)
- Toolbelt tools (WebView-based)

**Document but don't implement**: Guide users to Python API for these features.

---

## 9. Recommendations for it2 CLI

### P1 - Direct Implementation (High Value)

#### 1. Tab Color Commands (EASY)
```bash
it2 tab set-color <id> --rgb "R,G,B"
it2 tab clear-color <id>
```
Use `SetProfilePropertyRequest` with `KEY_TAB_COLOR`.

#### 2. Color Preset Management (MEDIUM)
```bash
it2 color-preset list
it2 color-preset get <name> [--format json|plist]
it2 color-preset export <name> <file>
it2 color-preset import <file>
it2 color-preset apply <name> <profile-guid>
```
Use `ColorPresetRequest` + file I/O.

#### 3. Enhanced Screen Capture (MEDIUM)
```bash
it2 get-buffer <id> --format text|json|styled
it2 get-buffer <id> --with-styles
it2 get-buffer <id> --range "line1:line2"
it2 get-buffer <id> --columns "col1:col2"
it2 get-buffer <id> --metadata-only
```
Leverage `include_styles`, `WindowedCoordRange`, metadata fields.

#### 4. Profile Property Bulk Operations (MEDIUM)
```bash
it2 profile set-properties <guid> --json '{...}'
it2 profile get-properties <guid> [key1 key2...]
it2 profile diff <guid1> <guid2>
it2 profile copy <source-guid> <dest-guid> [--keys key1,key2]
```
Use `SetProfilePropertyRequest.assignments` array for batching.

### P2 - Enhanced Functionality (Nice to Have)

#### 5. Invoke Function API (EASY)
```bash
it2 session stop-coprocess <id>
it2 session run-coprocess <id> <command> [--mute]
it2 tab select-pane <tab-id> --direction left|right|above|below
```
Use new `InvokeFunctionRequest`.

#### 6. Advanced Selection (MEDIUM)
```bash
it2 selection set <id> --mode word|line|box --range "x1,y1:x2,y2"
it2 selection get <id> --format text|json
```
Use `SelectionRequest` with `SubSelection`.

#### 7. Transaction Support for Batch Ops (EASY)
```bash
it2 batch begin
it2 session set-variable ... (multiple commands)
it2 batch end
```
Wrap operations with `TransactionRequest`.

#### 8. Window Property Management (EASY)
```bash
it2 window set-frame <id> --rect "x,y,w,h"
it2 window set-fullscreen <id> true|false
it2 window get-properties <id>
```
Use `SetPropertyRequest` / `GetPropertyRequest`.

### P3 - Advanced Features (Lower Priority)

#### 9. Variable Monitoring
```bash
it2 watch variable <scope> <name> [--filter expression]
```
Use `NotificationRequest` with `VariableMonitorRequest`.

#### 10. Keystroke Filtering
```bash
it2 filter-keys --pattern <pattern> --action <script>
```
Use `KeystrokeFilterRequest` (requires daemon).

#### 11. Menu Item Invocation
```bash
it2 menu invoke <identifier>
it2 menu query <identifier>
```
Use `MenuItemRequest`.

### Features to Document Only (Daemon Required)

- Custom RPC registration
- Status bar components
- Session title providers
- Context menu items
- Toolbelt tools
- Real-time variable watching (for automation)

---

## 10. Hidden Gems & Power User Features

### 1. Shell Prompt Intelligence
Location: `PTYSession.m:18545-18554`, `GetPromptResponse`

With shell integration:
- Extract prompt range, command range, output range
- Get working directory from prompt metadata
- Get command text
- Get exit status
- Track state: EDITING, RUNNING, FINISHED
- List all prompts with `ListPromptsRequest`

**Use Case**: Build command history analyzer, smart cd, command replay tools.

### 2. Broadcast Domains
Location: `proto/api.proto:458-467`, `BroadcastDomain`

- Group sessions for simultaneous input
- Must be in same window
- Must be disjoint (no session in multiple domains)
- Get current domains with `GetBroadcastDomainsRequest`
- Modify with `SetBroadcastDomainsRequest`

**Use Case**: CLI tool for managing broadcast groups.

### 3. Saved Arrangements Advanced Features
Location: `proto/api.proto:570-601`

- Save single window (not full arrangement)
- Restore into existing window as tabs
- List all arrangements
- Programmatic save/restore

**Use Case**: Session templates, workspace management.

### 4. Focus Change Events
Location: `proto/api.proto:1100-1134`

Detailed focus tracking:
- Application active/inactive
- Window became/resigned key
- Window is current (non-key terminal window)
- Selected tab changed
- Active session in tab changed

**Use Case**: Activity tracking, time tracking, context-aware automation.

### 5. tmux Integration API
Location: `proto/api.proto:400-456`

- List tmux connections
- Send raw tmux commands
- Control window visibility
- Create tmux windows programmatically
- Get connection ID and owning session

**Use Case**: Programmatic tmux control through iTerm2.

### 6. Profile Change Notifications
Location: `proto/api.proto:941-943`, `ProfileChangedNotification`

Monitor when profiles are modified:
```protobuf
message ProfileChangeRequest {
  optional string guid = 1;
}
```

**Use Case**: React to theme changes, profile updates.

### 7. Mouse Info Variable
Location: `iTermVariableScope+Session.m:337-354`

Incredibly detailed mouse event data:
```
mouseInfo = [x, y, button, count, modifiers, sideEffects, state]
```
- Position (x, y)
- Button number
- Click count (single, double, triple)
- Modifier keys array
- Side effects flag
- Mouse state

**Use Case**: Custom mouse handling, gesture detection.

---

## 11. Performance Best Practices

### From Source Code Analysis

1. **Use Transactions for > 10 operations**
   - Freezes app main loop
   - Ensures atomic updates
   - Keep transaction time < 100ms

2. **Request only needed columns**
   - Use `WindowedCoordRange.columns` to limit width
   - Saves network bandwidth and parsing time

3. **Disable styles if not needed**
   - `include_styles=false` reduces payload by ~70%

4. **Cache session/tab/window IDs**
   - Don't call `list_sessions` repeatedly
   - Subscribe to layout change notifications

5. **Use metadata-only properties**
   - `grid_size`, `buried`, `frame` don't require content transfer
   - Much faster than full buffer access

6. **Batch profile property changes**
   - `SetProfilePropertyRequest.assignments` array
   - Single request vs multiple

7. **Use specific session IDs**
   - Avoid "all" when possible
   - "active" is faster than list + filter

8. **Monitor connection health**
   - WebSocket can drop
   - Reconnection logic needed for long-running tools

---

## 12. Deprecation Warnings

### Deprecated Features Found

1. **Location Change Notification** (line 874, 995)
   - `NOTIFY_ON_LOCATION_CHANGE = 4 [deprecated=true]`
   - Use shell integration + prompt notifications instead

2. **Legacy Profile Keys**:
   - `KEY_DESCRIPTION` - No longer used
   - `KEY_CHILDREN` - Old bookmark organization
   - `KEY_DEPRECATED_BOOKMARKS` vs `KEY_NEW_BOOKMARKS`
   - `KEY_TERMINAL_PROFILE`, `KEY_KEYBOARD_PROFILE`, `KEY_DISPLAY_PROFILE`
   - `KEY_DEFAULT_BOOKMARK` - Use `KEY_DEFAULT_GUID`
   - `KEY_BONJOUR_*` - Bonjour support removed

3. **Color Handling**:
   - `KEY_DISABLE_BOLD` - Use `KEY_USE_BOLD_FONT` instead
   - `KEY_BACKGROUND_IMAGE_TILED_DEPRECATED` - Use `KEY_BACKGROUND_IMAGE_MODE`

4. **Text Processing**:
   - `KEY_TREAT_NON_ASCII_AS_DOUBLE_WIDTH` - Use `KEY_AMBIGUOUS_DOUBLE_WIDTH`
   - `KEY_USE_HFS_PLUS_MAPPING` - No longer relevant

5. **Display**:
   - `KEY_ANTI_ALIASING` - Split into ASCII and Non-ASCII variants
   - `KEY_FULLSCREEN` - Use window type instead

### Future-Proofing

- Always check status codes in responses
- Don't rely on deprecated notification types
- Use new profile keys when available
- Monitor iTerm2 release notes for API changes

---

## 13. Integration Plan for it2 CLI

### Week 1-2: Core Enhancements

**Day 1-3**: Tab Colors
- Implement `tab set-color`, `tab clear-color`, `tab get-color`
- Add RGB, hex, named color support
- Test with light/dark modes

**Day 4-7**: Color Presets
- List, get, export, import commands
- JSON and plist format support
- Apply preset to profile

**Day 8-10**: Screen Capture Enhanced
- Add `--with-styles` flag
- Implement column/line range filtering
- Add `--metadata-only` option

**Day 11-14**: Profile Operations
- Bulk property set/get
- Profile diff command
- Property copy between profiles

### Week 3-4: Advanced Features

**Day 15-18**: InvokeFunction API
- Coprocess management commands
- Pane navigation
- tmux integration commands

**Day 19-21**: Selection API
- Get/set selection with modes
- Multi-region selection support

**Day 22-24**: Transaction Support
- Batch command wrapper
- Automatic transaction detection for bulk ops

**Day 25-28**: Window Management
- Frame manipulation
- Fullscreen toggle
- Property queries

### Week 5-6: Polish & Documentation

**Day 29-35**: Testing
- Integration tests for each command
- Error handling edge cases
- Performance benchmarking

**Day 36-42**: Documentation
- Update README with new commands
- Write Python API vs CLI comparison guide
- Create migration examples

---

## 14. Questions & Answers

### Q1: Are there iTerm2 features in development that we should prepare for?

**Finding**: The `InvokeFunctionRequest` (proto line 104) appears to be a recent addition with extensible design. The method list in comments suggests this will be expanded. Recommend designing it2 CLI to easily add new invoke-function commands.

### Q2: What's the iTerm2 team's roadmap for API changes?

**Unable to determine from source** - Would need to check:
- GitHub issues labeled "api" or "python"
- Recent commits to `api.proto`
- Developer discussions

### Q3: Are there deprecation warnings we should know about?

**Yes - See Section 12 above**. Most importantly:
- Don't use `NOTIFY_ON_LOCATION_CHANGE`
- Avoid deprecated profile keys
- Use new color handling properties

### Q4: Any known performance bottlenecks in the WebSocket API?

**From source analysis**:
- Buffer requests with styles can be large (use sparingly)
- No explicit rate limiting, but app main loop can get backed up
- Transaction API exists specifically for performance during bulk operations
- WebSocket frame size limits not specified in source

**Recommendation**:
- Limit buffer requests to visible screen when possible
- Use transactions for > 10 operations
- Consider connection pooling for parallel operations

### Q5: Best practices for high-frequency operations?

**From source**:
1. Use variable change notifications instead of polling
2. Cache values that don't change often (IDs, grid size)
3. Use `trailing_lines` instead of full buffer for monitoring
4. Disable style information unless needed
5. Use specific session IDs, avoid "all" in loops
6. Batch property changes into single request

---

## 15. Code Examples

### Example 1: Efficient Tab Color Setting

```python
# GOOD - Direct profile property
await session.async_set_profile_property("Tab Color", {
    "Red Component": 1.0,
    "Green Component": 0.5,
    "Blue Component": 0.0,
    "Color Space": "sRGB"
})
await session.async_set_profile_property("Use Tab Color", True)

# BAD - Indirect variable manipulation (doesn't work well)
# Don't try to set via variables - no direct variable exposed!
```

### Example 2: Bulk Profile Updates

```python
# Use assignments array for atomic update
assignments = [
    {"key": "Foreground Color", "json_value": json.dumps(color_dict)},
    {"key": "Background Color", "json_value": json.dumps(bg_dict)},
    {"key": "Use Bold Color", "json_value": "true"},
]

request = iterm2.SetProfilePropertyRequest()
request.session = session_id
request.assignments = assignments
await connection.async_send_request(request)
```

### Example 3: Efficient Buffer Monitoring

```python
# Get just last 10 lines without styles
request = iterm2.GetBufferRequest()
request.session = session_id
request.line_range.trailing_lines = 10
request.include_styles = False

response = await connection.async_get_buffer(request)
# Much faster than full buffer!
```

### Example 4: Using Transactions

```python
# Freeze time for bulk updates
await connection.async_transaction(begin=True)
try:
    for session in sessions:
        await session.async_set_variable("user.status", "updating")
        await session.async_set_profile_property("Badge Text", "UPDATING")
finally:
    await connection.async_transaction(begin=False)
```

---

## Summary

This analysis of the iTerm2 source code reveals:

1. **33 RPC methods** - Several undocumented or under-documented
2. **80+ variables** across 4 scopes - Complete catalog provided
3. **28 color keys** - Full preset format documented
4. **100+ profile properties** - Comprehensive list with types
5. **Performance APIs** - Transactions, windowed ranges, metadata queries
6. **Advanced buffer APIs** - Styles, URLs, images, run-length encoding
7. **Direct tab color API** - Much simpler than variable approach
8. **Plugin system** - Not suitable for CLI, document only

**Key Recommendations**:
- Implement P1 features (tab colors, presets, enhanced capture, bulk profiles)
- Use direct APIs where available (tab color via profile property)
- Leverage performance features (transactions, column windows, metadata-only)
- Document but don't implement daemon-required features
- Follow deprecation guidance

**Estimated Development Time**: 6-8 weeks for full P1/P2 implementation

This should provide everything needed to significantly enhance it2 CLI's Python API parity!
