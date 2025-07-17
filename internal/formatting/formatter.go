package formatting

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/it2/internal/client"
	pb "github.com/tmc/it2/proto"
	"gopkg.in/yaml.v3"
)

type Formatter struct {
	format string
}

func New(format string) *Formatter {
	return &Formatter{format: format}
}

func (f *Formatter) FormatSessions(sessions []*client.SessionInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(sessions)
	case "yaml":
		return f.formatYAML(sessions)
	default:
		return f.formatText(sessions)
	}
}

func (f *Formatter) FormatTabResponse(resp *pb.CreateTabResponse) error {
	if resp.GetTabId() != 0 {
		fmt.Printf("Created tab with ID: %d\n", resp.GetTabId())
	}
	if resp.GetWindowId() != "" {
		fmt.Printf("Window ID: %s\n", resp.GetWindowId())
	}
	if resp.GetSessionId() != "" {
		fmt.Printf("Session ID: %s\n", resp.GetSessionId())
	}
	return nil
}

func (f *Formatter) FormatBuffer(resp *pb.GetBufferResponse) error {
	for _, line := range resp.GetContents() {
		if line.GetText() != "" {
			fmt.Print(line.GetText())
		}
	}
	return nil
}

func (f *Formatter) FormatJobs(jobs []*client.JobInfo) error {
	switch f.format {
	case "json":
		return f.formatJSON(jobs)
	case "yaml":
		return f.formatYAML(jobs)
	default:
		return f.formatJobsText(jobs)
	}
}

func (f *Formatter) formatJobsText(jobs []*client.JobInfo) error {
	if len(jobs) == 0 {
		fmt.Println("No running jobs found")
		return nil
	}

	for _, job := range jobs {
		fmt.Printf("Job %s: %s - %s\n", job.JobID, job.Status, job.Command)
	}
	return nil
}

func (f *Formatter) formatText(sessions []*client.SessionInfo) error {
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	for _, session := range sessions {
		fmt.Printf("Session ID: %s\n", session.SessionID)
		if session.WindowID != "" {
			fmt.Printf("  Window ID: %s\n", session.WindowID)
		}
		if session.TabID != "" {
			fmt.Printf("  Tab ID: %s\n", session.TabID)
		}
		if session.SessionName != "" {
			fmt.Printf("  Name: %s\n", session.SessionName)
		}
		fmt.Println(strings.Repeat("-", 40))
	}
	return nil
}

func (f *Formatter) formatJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (f *Formatter) formatYAML(v interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(v)
}