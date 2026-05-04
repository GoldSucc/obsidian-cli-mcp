package tools

import (
	"context"
	"strconv"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TaskInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Ref    string `json:"ref,omitempty" jsonschema:"task reference as path:line"`
	File   string `json:"file,omitempty" jsonschema:"file name resolved like a wikilink (no path/extension)"`
	Path   string `json:"path,omitempty" jsonschema:"exact path from vault root"`
	Line   int    `json:"line,omitempty" jsonschema:"line number of the task within the file"`
	Toggle bool   `json:"toggle,omitempty" jsonschema:"toggle the task status"`
	Done   bool   `json:"done,omitempty" jsonschema:"mark task as done"`
	Todo   bool   `json:"todo,omitempty" jsonschema:"mark task as todo"`
	Daily  bool   `json:"daily,omitempty" jsonschema:"target the daily note"`
	Status string `json:"status,omitempty" jsonschema:"single-character status to set"`
}

func taskHandler(ctx context.Context, _ *mcp.CallToolRequest, in TaskInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Ref != "" {
		params["ref"] = in.Ref
	}
	if in.File != "" {
		params["file"] = in.File
	}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Line != 0 {
		params["line"] = strconv.Itoa(in.Line)
	}
	if in.Status != "" {
		params["status"] = in.Status
	}
	flags := []string{}
	if in.Toggle {
		flags = append(flags, "toggle")
	}
	if in.Done {
		flags = append(flags, "done")
	}
	if in.Todo {
		flags = append(flags, "todo")
	}
	if in.Daily {
		flags = append(flags, "daily")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "task", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TasksInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	File    string `json:"file,omitempty" jsonschema:"file name resolved like a wikilink (no path/extension)"`
	Path    string `json:"path,omitempty" jsonschema:"exact path from vault root"`
	Total   bool   `json:"total,omitempty" jsonschema:"include only the total count"`
	Done    bool   `json:"done,omitempty" jsonschema:"include only completed tasks"`
	Todo    bool   `json:"todo,omitempty" jsonschema:"include only todo tasks"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra task details"`
	Active  bool   `json:"active,omitempty" jsonschema:"include only active (non-cancelled) tasks"`
	Daily   bool   `json:"daily,omitempty" jsonschema:"target the daily note"`
	Status  string `json:"status,omitempty" jsonschema:"single-character status to filter on"`
	Format  string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func tasksHandler(ctx context.Context, _ *mcp.CallToolRequest, in TasksInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.File != "" {
		params["file"] = in.File
	}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Status != "" {
		params["status"] = in.Status
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Done {
		flags = append(flags, "done")
	}
	if in.Todo {
		flags = append(flags, "todo")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	if in.Active {
		flags = append(flags, "active")
	}
	if in.Daily {
		flags = append(flags, "daily")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "tasks", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterTasks(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_task",
		Description: "Show or update a single task. Identify via ref (path:line) or file/path plus line. Toggle, set done/todo, or status=<char>.",
	}, taskHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tasks",
		Description: "List tasks across vault or within a file. Filters: done, todo, active, status, daily. Format json/tsv/csv.",
	}, tasksHandler)
}
