package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FolderInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Path  string `json:"path" jsonschema:"folder path from vault root (required)"`
	Info  string `json:"info,omitempty" jsonschema:"detail to return: files, folders, or size"`
}

func folderHandler(ctx context.Context, _ *mcp.CallToolRequest, in FolderInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Info != "" {
		params["info"] = in.Info
	}
	out, err := exec.Run(ctx, exec.Args{Command: "folder", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type FoldersInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Folder string `json:"folder,omitempty" jsonschema:"folder path to list under; omit for vault root"`
	Total  bool   `json:"total,omitempty" jsonschema:"include total count"`
}

func foldersHandler(ctx context.Context, _ *mcp.CallToolRequest, in FoldersInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Folder != "" {
		params["folder"] = in.Folder
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "folders", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterFolders(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_folder",
		Description: "Get information about a folder. Use info=files|folders|size for specific detail.",
	}, folderHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_folders",
		Description: "List folders in the vault or under a specific folder. Use total to include count.",
	}, foldersHandler)
}
