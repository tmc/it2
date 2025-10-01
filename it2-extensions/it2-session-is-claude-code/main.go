package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
)

//go:embed it2-session-is-claude-code.sh
var script string

func main() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	args := append([]string{"-s"}, os.Args[1:]...)
	cmd := exec.Command(shell, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// Write script to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating stdin pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting command: %v\n", err)
		os.Exit(1)
	}

	if _, err := stdin.Write([]byte(script)); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing script: %v\n", err)
		os.Exit(1)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error waiting for command: %v\n", err)
		os.Exit(1)
	}
}
