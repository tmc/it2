# Session Tail Design

## Current Implementation (Polling)

The `it2 session tail` command currently uses a polling-based approach:

1. Fetches last 100 lines of buffer every interval (default 1s)
2. Compares content with previous fetch using string comparison
3. Extracts and displays only new content
4. Handles edge cases like scrollback overflow and screen clears

### Advantages
- Simple to implement
- Works with existing API
- No persistent connection needed

### Disadvantages
- Wasteful - polls even when no changes
- Latency - updates only every polling interval
- Scales poorly with many concurrent tail sessions

## Future Enhancement: Notification-Based Tail

iTerm2's API supports `NOTIFY_ON_SCREEN_UPDATE` notifications that provide a much more efficient approach.

### Notification Protocol

Based on iTerm2's `api.proto`:

```protobuf
enum NotificationType {
  NOTIFY_ON_SCREEN_UPDATE = 2;
}

message NotificationRequest {
  optional string session = 1;  // Session to monitor
  optional bool subscribe = 2;   // true to subscribe, false to unsubscribe
  optional NotificationType notification_type = 3;
}

message ScreenUpdateNotification {
  optional string session = 1;  // Session that was updated
}
```

### Proposed Implementation

1. **Subscribe** - Send NotificationRequest with:
   - `session` = target session ID
   - `subscribe` = true
   - `notification_type` = NOTIFY_ON_SCREEN_UPDATE

2. **Listen** - Keep WebSocket connection open and listen for notifications

3. **Fetch on Notification** - When ScreenUpdateNotification received:
   - Track last cursor position (Coord.y)
   - Use `trailing_lines` or `windowed_coord_range` to fetch only new lines
   - Display incremental content

4. **Unsubscribe** - On exit/Ctrl+C, send NotificationRequest with `subscribe` = false

### Advantages
- Event-driven - only fetches when content changes
- Lower latency - immediate updates
- More efficient - no wasted polling
- Better scalability - server pushes updates

### Coordinate Tracking

iTerm2 uses absolute coordinates that remain stable:
- Each line has a fixed `y` coordinate that never changes
- As scrollback fills, coordinates increase but existing lines stay stable
- Track `last_seen_y` and fetch lines with `y > last_seen_y`

### Implementation Tasks

- [ ] Add notification subscription to client
- [ ] Implement notification message handling
- [ ] Add coordinate-based tracking
- [ ] Update tail command to use notifications
- [ ] Add fallback to polling if notifications fail
- [ ] Update documentation and examples

## References

- iTerm2 proto: `/Volumes/tmc/go/src/github.com/gnachman/iTerm2/proto/api.proto`
- Notification types: Lines 869-887
- NotificationRequest: Lines 955-976
- ScreenUpdateNotification: Lines 1054-1056
