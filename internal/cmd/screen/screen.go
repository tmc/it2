package screen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tmc/it2/internal/cmdutil"
	"github.com/tmc/it2/internal/formatting"
	pb "github.com/tmc/it2/proto"
)

// NewCommand creates the screen command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen",
		Short: "Screen capture and recording utilities",
		Long:  "Commands for taking screenshots, recording screens, and exporting captures",
	}

	cmd.AddCommand(newCaptureCommand())
	cmd.AddCommand(newRecordCommand())
	cmd.AddCommand(newExportCommand())

	return cmd
}

func newCaptureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capture [session-id]",
		Short: "Take a screenshot of a session",
		Long:  "Take a screenshot of the specified session or the active session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			output, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")
			quality, _ := cmd.Flags().GetFloat64("quality")
			transparent, _ := cmd.Flags().GetBool("transparent")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to trigger screen capture
			// This uses iTerm2's built-in screenshot functionality
			invocation := fmt.Sprintf(`
				set theSession to current session of current tab of current window
				if "%s" is not "active" then
					set sessionId to "%s"
					-- Find session by ID if not active
				end if

				set outputPath to "%s"
				if outputPath is "" then
					set outputPath to (path to desktop as string) & "iterm2_capture_" & (do shell script "date +%%Y%%m%%d_%%H%%M%%S") & ".%s"
				end if

				tell theSession
					set screenCapture to capture screen as %s
					if %v then
						set screenCapture to screenCapture with transparency
					end if
					save screenCapture to file outputPath
				end tell

				return outputPath
			`, sessionID, sessionID, output, format, format, transparent)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_App{
					App: &pb.InvokeFunctionRequest_App{},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to capture screen: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("screen capture failed: %s", response.GetError().GetErrorReason())
			}

			if success := response.GetSuccess(); success != nil {
				fmt.Printf("Screenshot saved to: %s\n", success.GetJsonResult())
			} else {
				fmt.Println("Screenshot captured successfully")
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: Desktop/iterm2_capture_TIMESTAMP.format)")
	cmd.Flags().StringP("format", "f", "png", "Output format (png, jpg, tiff)")
	cmd.Flags().Float64P("quality", "q", 1.0, "Image quality for JPEG (0.0-1.0)")
	cmd.Flags().BoolP("transparent", "t", false, "Include transparency (PNG only)")

	return cmd
}

func newRecordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Start or stop screen recording",
		Long:  "Start or stop screen recording of iTerm2 sessions",
	}

	cmd.AddCommand(newRecordStartCommand())
	cmd.AddCommand(newRecordStopCommand())

	return cmd
}

func newRecordStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [session-id]",
		Short: "Start screen recording",
		Long:  "Start recording the screen of the specified session or active session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			output, _ := cmd.Flags().GetString("output")
			fps, _ := cmd.Flags().GetInt("fps")
			duration, _ := cmd.Flags().GetDuration("duration")

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Generate default output path if not specified
			if output == "" {
				timestamp := time.Now().Format("20060102_150405")
				output = filepath.Join(os.Getenv("HOME"), "Desktop", fmt.Sprintf("iterm2_recording_%s.mov", timestamp))
			}

			// Use InvokeFunction to start recording
			invocation := fmt.Sprintf(`
				set theSession to current session of current tab of current window
				if "%s" is not "active" then
					set sessionId to "%s"
					-- Find session by ID if not active
				end if

				set outputPath to "%s"
				set frameRate to %d
				set maxDuration to %f

				tell theSession
					start recording to file outputPath with frame rate frameRate and duration maxDuration
				end tell

				return "Recording started: " & outputPath
			`, sessionID, sessionID, output, fps, duration.Seconds())

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to start recording: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("recording start failed: %s", response.GetError().GetErrorReason())
			}

			if success := response.GetSuccess(); success != nil {
				fmt.Println(success.GetJsonResult())
			} else {
				fmt.Printf("Recording started, saving to: %s\n", output)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: Desktop/iterm2_recording_TIMESTAMP.mov)")
	cmd.Flags().IntP("fps", "f", 30, "Frames per second (10-60)")
	cmd.Flags().DurationP("duration", "d", 0, "Maximum recording duration (0 = unlimited)")

	return cmd
}

func newRecordStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [session-id]",
		Short: "Stop screen recording",
		Long:  "Stop the current screen recording for the specified session or active session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := "active"
			if len(args) > 0 {
				sessionID = args[0]
			}

			wsURL, timeout, _ := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Use InvokeFunction to stop recording
			invocation := fmt.Sprintf(`
				set theSession to current session of current tab of current window
				if "%s" is not "active" then
					set sessionId to "%s"
					-- Find session by ID if not active
				end if

				tell theSession
					stop recording
				end tell

				return "Recording stopped"
			`, sessionID, sessionID)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to stop recording: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("recording stop failed: %s", response.GetError().GetErrorReason())
			}

			if success := response.GetSuccess(); success != nil {
				fmt.Println(success.GetJsonResult())
			} else {
				fmt.Println("Recording stopped")
			}

			return nil
		},
	}

	return cmd
}

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <format> [session-id]",
		Short: "Export capture in specified format",
		Long:  "Export the last screen capture in the specified format",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			exportFormat := args[0]
			sessionID := "active"
			if len(args) > 1 {
				sessionID = args[1]
			}

			output, _ := cmd.Flags().GetString("output")
			quality, _ := cmd.Flags().GetFloat64("quality")

			wsURL, timeout, format := cmdutil.GetFlags(cmd)

			ctx, cancel := cmdutil.CreateContext(timeout)
			defer cancel()

			c, err := cmdutil.ConnectClient(ctx, wsURL)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Validate export format
			validFormats := []string{"png", "jpg", "jpeg", "tiff", "pdf", "svg"}
			valid := false
			for _, vf := range validFormats {
				if exportFormat == vf {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("unsupported format: %s. Valid formats: %v", exportFormat, validFormats)
			}

			// Generate default output path if not specified
			if output == "" {
				timestamp := time.Now().Format("20060102_150405")
				output = filepath.Join(os.Getenv("HOME"), "Desktop", fmt.Sprintf("iterm2_export_%s.%s", timestamp, exportFormat))
			}

			// For this implementation, we'll get the buffer content and format it
			// In a real implementation, this might use more sophisticated export capabilities
			if exportFormat == "text" || exportFormat == "txt" {
				// Export as text
				buffer, err := c.GetBuffer(ctx, sessionID, nil, false)
				if err != nil {
					return fmt.Errorf("failed to get buffer: %w", err)
				}

				var text string
				for _, line := range buffer.GetContents() {
					text += line.GetText() + "\n"
				}

				err = os.WriteFile(output, []byte(text), 0644)
				if err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				fmt.Printf("Text exported to: %s\n", output)
				return nil
			}

			// For image formats, use the capture command with export
			invocation := fmt.Sprintf(`
				set theSession to current session of current tab of current window
				if "%s" is not "active" then
					set sessionId to "%s"
					-- Find session by ID if not active
				end if

				set outputPath to "%s"
				set exportFormat to "%s"
				set imageQuality to %f

				tell theSession
					set screenCapture to capture screen
					save screenCapture to file outputPath as %s with quality imageQuality
				end tell

				return "Exported to: " & outputPath
			`, sessionID, sessionID, output, exportFormat, quality, exportFormat)

			request := &pb.InvokeFunctionRequest{
				Context: &pb.InvokeFunctionRequest_Session{
					Session: &pb.InvokeFunctionRequest_Session{
						SessionId: sessionID,
					},
				},
				Invocation: invocation,
				Timeout:    float64(timeout.Seconds()),
			}

			response, err := c.InvokeFunction(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to export: %w", err)
			}

			if response.GetError() != nil {
				return fmt.Errorf("export failed: %s", response.GetError().GetErrorReason())
			}

			formatter := formatting.New(format)
			if success := response.GetSuccess(); success != nil {
				if format == "json" {
					return formatter.FormatGeneric(success)
				}
				fmt.Println(success.GetJsonResult())
			} else {
				fmt.Printf("Exported to: %s\n", output)
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: Desktop/iterm2_export_TIMESTAMP.format)")
	cmd.Flags().Float64P("quality", "q", 1.0, "Image quality for lossy formats (0.0-1.0)")

	return cmd
}
