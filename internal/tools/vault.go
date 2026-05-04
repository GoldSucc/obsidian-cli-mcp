package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type VaultInfoInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Info  string `json:"info,omitempty" jsonschema:"detail to return: name, path, files, folders, or size"`
}

func vaultInfoHandler(ctx context.Context, _ *mcp.CallToolRequest, in VaultInfoInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Info != "" {
		params["info"] = in.Info
	}
	out, err := exec.Run(ctx, exec.Args{Command: "vault", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type VaultsInput struct {
	Total   bool `json:"total,omitempty" jsonschema:"include total count"`
	Verbose bool `json:"verbose,omitempty" jsonschema:"include verbose details for each vault"`
}

func vaultsHandler(ctx context.Context, _ *mcp.CallToolRequest, in VaultsInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "vaults", Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type ReloadInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
}

func reloadHandler(ctx context.Context, _ *mcp.CallToolRequest, in ReloadInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "reload", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type RestartInput struct{}

func restartHandler(ctx context.Context, _ *mcp.CallToolRequest, in RestartInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "restart"})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type VersionInput struct{}

func versionHandler(ctx context.Context, _ *mcp.CallToolRequest, in VersionInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "version"})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterVault(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_vault_info",
		Description: "Get information about a vault. Use info=name|path|files|folders|size for specific detail.",
	}, vaultInfoHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_vaults",
		Description: "List configured vaults. Use total for count, verbose for details.",
	}, vaultsHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_reload",
		Description: "Reload the vault.",
	}, reloadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_restart",
		Description: "Restart the Obsidian app.",
	}, restartHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_version",
		Description: "Show the Obsidian version.",
	}, versionHandler)
}
