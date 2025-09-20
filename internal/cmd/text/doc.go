// Package text provides commands for text operations in iTerm2.
//
// Text operations allow interaction with the terminal content, including
// sending input and retrieving output. This package provides commands to:
//
//   - send: Send text input to sessions (keystrokes, commands)
//   - buffer: Retrieve terminal buffer contents (screen or scrollback)
//   - select: Control text selection within terminals
//
// Text operations are fundamental for terminal automation, allowing
// programs to interact with running processes and extract information
// from terminal output.
//
// Examples:
//   it2 text send w0t0p0s0 "echo hello\n"
//   it2 text buffer w0t0p0s0 --lines 100
//   it2 text select w0t0p0s0 --start 0,0 --end 10,5
package text