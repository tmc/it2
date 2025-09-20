// Package job provides commands for job monitoring in iTerm2.
//
// Jobs represent running processes within iTerm2 sessions. This package
// provides commands to:
//
//   - list: List all running jobs in sessions
//   - monitor: Watch job status changes in real-time
//
// Job monitoring allows tracking of process execution within terminals,
// useful for automation and monitoring workflows. Jobs are identified
// by their process IDs and associated session information.
//
// Examples:
//   it2 job list w0t0p0s0
//   it2 job monitor w0t0p0s0 --watch
package job