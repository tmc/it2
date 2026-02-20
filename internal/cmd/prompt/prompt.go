package prompt

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdcore"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the prompt command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prompt",
		GroupID: "monitoring",
		Short:   "Manage prompts and command history in iTerm2 sessions",
		Long:    "Commands for listing, monitoring, and searching prompts and command history in iTerm2 sessions",
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newMonitorCommand())
	cmd.AddCommand(newSearchCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [<session-id>]",
		Short: "List historical prompts in a session",
		Long:  "List all historical prompt unique IDs in a session. If no session-id is provided, uses $ITERM_SESSION_ID environment variable. Requires Shell Integration to be enabled in iTerm2.",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sessionID string
			if len(args) > 0 {
				sessionID = args[0]
			}

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			sessionID, err = c.ResolveSessionID(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to resolve session ID: %w", err)
			}

			response, err := c.ListPrompts(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to list prompts: %w", err)
			}

			switch response.GetStatus() {
			case pb.ListPromptsResponse_OK:
				prompts := response.GetUniquePromptId()
				if len(prompts) == 0 {
					fmt.Println("No prompts found in session.")
					fmt.Println("Make sure Shell Integration is enabled and you have run some commands.")
					return nil
				}

				// Check if user wants quiet mode (IDs only)
				quiet, _ := cmd.Flags().GetBool("quiet")

				if !quiet {
					// Fetch details for each prompt and display in table
					fmt.Printf("Found %d prompt(s) in session %s:\n\n", len(prompts), sessionID)

					// Table header
					fmt.Printf("%-4s %-40s %-40s %-25s %-10s %-8s\n", "#", "Prompt ID", "Command", "Directory", "Exit Code", "State")
					fmt.Println(strings.Repeat("-", 131))

					for i, promptID := range prompts {
						// Get prompt details using the client's GetPromptByID method
						var command, directory, exitCode, state string = "-", "-", "-", "-"

						// Actually fetch details now
						promptResp, err := c.GetPromptByID(ctx, sessionID, promptID)
						if err != nil || promptResp == nil || promptResp.GetStatus() != pb.GetPromptResponse_OK {
							// Error or non-OK status - show defaults
						} else {
							// OK status - extract all data
							if promptResp.GetCommand() != "" {
								command = promptResp.GetCommand()
								// Truncate long commands
								if len(command) > 38 {
									command = command[:35] + "..."
								}
							}

							// Get working directory
							if promptResp.GetWorkingDirectory() != "" {
								directory = promptResp.GetWorkingDirectory()
								// Shorten home directory path
								if home := os.Getenv("HOME"); home != "" && strings.HasPrefix(directory, home) {
									directory = "~" + directory[len(home):]
								}
								// Truncate long paths
								if len(directory) > 23 {
									// Keep the last part of the path
									parts := strings.Split(directory, "/")
									if len(parts) > 2 {
										directory = ".../" + parts[len(parts)-1]
									} else {
										directory = directory[:20] + "..."
									}
								}
							}

							// Get state
							switch promptResp.GetPromptState() {
							case pb.GetPromptResponse_EDITING:
								state = "editing"
								exitCode = "-"
							case pb.GetPromptResponse_RUNNING:
								state = "running"
								exitCode = "-"
							case pb.GetPromptResponse_FINISHED:
								state = "done"
								// Get exit status for finished commands
								if promptResp.GetExitStatus() != 0 {
									exitCode = fmt.Sprintf("%d", promptResp.GetExitStatus())
								} else {
									exitCode = "0"
								}
							}
						}

						// Show full prompt ID - they're UUIDs so always 36 chars
						fmt.Printf("%-4d %-40s %-40s %-25s %-10s %-8s\n",
							i+1, promptID, command, directory, exitCode, state)
					}
				} else {
					// Quiet mode - just output prompt IDs to stdout
					for _, promptID := range prompts {
						fmt.Println(promptID)
					}
				}

			case pb.ListPromptsResponse_SESSION_NOT_FOUND:
				return fmt.Errorf("session not found: %s", sessionID)
			default:
				return fmt.Errorf("unexpected response status: %v", response.GetStatus())
			}

			return nil
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Quiet mode - output only prompt IDs to stdout")

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <prompt-id>",
		Short: "Get specific prompt details by ID",
		Long:  "Get detailed information about a specific prompt using its unique ID. Requires Shell Integration to be enabled in iTerm2.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			promptID := args[0]

			_, timeout, format := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Try to get prompt without session ID first
			promptReq := &pb.GetPromptRequest{
				UniquePromptId: &promptID,
			}

			msg := &pb.ClientOriginatedMessage{
				Submessage: &pb.ClientOriginatedMessage_GetPromptRequest{
					GetPromptRequest: promptReq,
				},
			}

			response, err := c.SendRequest(ctx, msg)
			if err != nil {
				return fmt.Errorf("failed to get prompt: %w", err)
			}

			promptResp := response.GetGetPromptResponse()
			if promptResp == nil {
				return fmt.Errorf("no prompt response received")
			}

			switch promptResp.GetStatus() {
			case pb.GetPromptResponse_OK:
				return displayPromptDetails(promptResp, format)
			case pb.GetPromptResponse_SESSION_NOT_FOUND:
				return fmt.Errorf("session not found for prompt")
			case pb.GetPromptResponse_PROMPT_UNAVAILABLE:
				return fmt.Errorf("prompt not found or unavailable: %s", promptID)
			case pb.GetPromptResponse_REQUEST_MALFORMED:
				return fmt.Errorf("request malformed")
			default:
				return fmt.Errorf("unexpected response status: %v", promptResp.GetStatus())
			}
		},
	}

	return cmd
}

func newMonitorCommand() *cobra.Command {
	var modes []string

	cmd := &cobra.Command{
		Use:   "monitor <session-id>",
		Short: "Monitor prompt events in real-time",
		Long:  "Monitor prompt notifications in real-time. Shows new prompts, command starts, and command completions. Requires Shell Integration to be enabled in iTerm2.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			_, timeout, _ := cmdcore.GetFlags(cmd)

			// Create context that can be canceled but don't set timeout for monitoring
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			connectCtx, connectCancel := cmdcore.CreateContext(timeout)
			c, err := cmdcore.ConnectClient(connectCtx)
			connectCancel()
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Start monitoring
			ch, err := c.MonitorSession(ctx, sessionID, modes)
			if err != nil {
				return fmt.Errorf("failed to start prompt monitoring: %w", err)
			}

			fmt.Printf("Monitoring prompt events for session %s...\n", sessionID)
			fmt.Printf("Press Ctrl+C to stop monitoring.\n\n")

			for notification := range ch {
				displayPromptNotification(notification)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&modes, "modes", []string{"prompt", "command_start", "command_end"}, "Monitor modes: prompt, command_start, command_end")

	return cmd
}

func newSearchCommand() *cobra.Command {
	var pattern string
	var regex bool

	cmd := &cobra.Command{
		Use:   "search <session-id> <pattern>",
		Short: "Search through command history",
		Long:  "Search through command history by filtering prompts based on command text. Requires Shell Integration to be enabled in iTerm2.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			pattern = args[1]

			_, timeout, _ := cmdcore.GetFlags(cmd)
			ctx, cancel := cmdcore.CreateContext(timeout)
			defer cancel()

			c, err := cmdcore.ConnectClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// First, list all prompts
			listResponse, err := c.ListPrompts(ctx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to list prompts: %w", err)
			}

			if listResponse.GetStatus() != pb.ListPromptsResponse_OK {
				return fmt.Errorf("failed to list prompts: %v", listResponse.GetStatus())
			}

			prompts := listResponse.GetUniquePromptId()
			if len(prompts) == 0 {
				fmt.Println("No prompts found in session.")
				return nil
			}

			// Compile regex if needed
			var re *regexp.Regexp
			if regex {
				var err error
				re, err = regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid regex pattern: %w", err)
				}
			}

			fmt.Printf("Searching %d prompt(s) for pattern: %s\n", len(prompts), pattern)
			fmt.Println(strings.Repeat("-", 60))

			matchCount := 0
			for _, promptID := range prompts {
				promptResponse, err := c.GetPromptByID(ctx, sessionID, promptID)
				if err != nil {
					continue // Skip prompts that can't be retrieved
				}

				if promptResponse.GetStatus() != pb.GetPromptResponse_OK {
					continue
				}

				command := promptResponse.GetCommand()
				if command == "" {
					continue
				}

				var matches bool
				if regex {
					matches = re.MatchString(command)
				} else {
					matches = strings.Contains(command, pattern)
				}

				if matches {
					matchCount++
					fmt.Printf("Match %d (ID: %s):\n", matchCount, promptID)
					fmt.Printf("  Command: %s\n", command)
					if workingDir := promptResponse.GetWorkingDirectory(); workingDir != "" {
						fmt.Printf("  Directory: %s\n", workingDir)
					}
					if promptResponse.GetPromptState() == pb.GetPromptResponse_FINISHED {
						fmt.Printf("  Exit Status: %d\n", promptResponse.GetExitStatus())
					}
					fmt.Println()
				}
			}

			fmt.Printf("Found %d matching command(s).\n", matchCount)
			return nil
		},
	}

	cmd.Flags().BoolVar(&regex, "regex", false, "Use regex pattern matching")

	return cmd
}

func displayPromptDetails(response *pb.GetPromptResponse, format string) error {
	if format == "json" {
		return formatting.PrintJSON(response)
	}

	fmt.Printf("Prompt Details:\n")
	fmt.Println(strings.Repeat("-", 40))

	if promptID := response.GetUniquePromptId(); promptID != "" {
		fmt.Printf("ID: %s\n", promptID)
	}

	if command := response.GetCommand(); command != "" {
		fmt.Printf("Command: %s\n", command)
	}

	if workingDir := response.GetWorkingDirectory(); workingDir != "" {
		fmt.Printf("Working Directory: %s\n", workingDir)
	}

	state := response.GetPromptState()
	var stateStr string
	switch state {
	case pb.GetPromptResponse_EDITING:
		stateStr = "EDITING"
	case pb.GetPromptResponse_RUNNING:
		stateStr = "RUNNING"
	case pb.GetPromptResponse_FINISHED:
		stateStr = "FINISHED"
	default:
		stateStr = "UNKNOWN"
	}
	fmt.Printf("State: %s\n", stateStr)

	if state == pb.GetPromptResponse_FINISHED {
		fmt.Printf("Exit Status: %d\n", response.GetExitStatus())
	}

	if promptRange := response.GetPromptRange(); promptRange != nil {
		fmt.Printf("Prompt Range: Line %d-%d, Column %d-%d\n",
			promptRange.GetStart().GetY(), promptRange.GetEnd().GetY(),
			promptRange.GetStart().GetX(), promptRange.GetEnd().GetX())
	}

	if commandRange := response.GetCommandRange(); commandRange != nil {
		fmt.Printf("Command Range: Line %d-%d, Column %d-%d\n",
			commandRange.GetStart().GetY(), commandRange.GetEnd().GetY(),
			commandRange.GetStart().GetX(), commandRange.GetEnd().GetX())
	}

	if outputRange := response.GetOutputRange(); outputRange != nil {
		fmt.Printf("Output Range: Line %d-%d, Column %d-%d\n",
			outputRange.GetStart().GetY(), outputRange.GetEnd().GetY(),
			outputRange.GetStart().GetX(), outputRange.GetEnd().GetX())
	}

	return nil
}

func displayPromptNotification(notification *pb.PromptNotification) {
	timestamp := time.Now().Format("15:04:05")
	session := notification.GetSession()

	switch event := notification.Event.(type) {
	case *pb.PromptNotification_Prompt:
		fmt.Printf("[%s] PROMPT (session: %s)\n", timestamp, session)
		if prompt := event.Prompt.GetPrompt(); prompt != nil {
			if cmd := prompt.GetCommand(); cmd != "" {
				fmt.Printf("  Command: %s\n", cmd)
			}
			if wd := prompt.GetWorkingDirectory(); wd != "" {
				fmt.Printf("  Directory: %s\n", wd)
			}
			if promptID := prompt.GetUniquePromptId(); promptID != "" {
				fmt.Printf("  ID: %s\n", promptID)
			}
		}
		fmt.Println()

	case *pb.PromptNotification_CommandStart:
		fmt.Printf("[%s] COMMAND_START (session: %s)\n", timestamp, session)
		if cmd := event.CommandStart.GetCommand(); cmd != "" {
			fmt.Printf("  Command: %s\n", cmd)
		}
		fmt.Println()

	case *pb.PromptNotification_CommandEnd:
		fmt.Printf("[%s] COMMAND_END (session: %s)\n", timestamp, session)
		fmt.Printf("  Exit Status: %d\n", event.CommandEnd.GetStatus())
		fmt.Println()
	}
}
