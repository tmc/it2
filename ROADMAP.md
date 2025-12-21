# it2 CLI - Roadmap

## Active Development
- [ ] Documentation Cleanup
  - [ ] Rename `CLAUDE_EXTENSIONS_REFERENCE.md` to `CLAUDE_PLUGINS_REFERENCE.md`
  - [ ] Update README to link to new docs
- [ ] Plugin System Refactor
  - [ ] Integrate plugin hook (`internal/cmd/plugins/claude_code_hook.go`)
  - [ ] Add `--plugin` flag to `it2-session-monitor`
  - [ ] Add colorized output
  - [ ] Implement stats/summary modes

## Upcoming Features
- [ ] Align `it2 session splits` tree output
- [ ] Shared `--url` flag handling
- [ ] Implement `it2 recordings` API wrappers
- [ ] Session search/filter - `it2 session list --filter "name~build"`
- [ ] InvokeFunction API wrapper
- [ ] Transaction API for batch operations
- [ ] Notification watch command
- [ ] Window property management (frame, fullscreen)
- [ ] Enhanced buffer capture (styles, colors, links)
- [ ] Multi-session operations
- [ ] Open up client apis/packages for library use

## Experimental / Maybe
- [ ] Semantic History integration for OSC 8 links
- [ ] TUI snapshot mode for tail
- [ ] Tab/window focused commands
- [ ] Color preset management
- [ ] Session templates (save/restore)
- [ ] Export/import session state

## Future
- [ ] Session recording (record/replay)

