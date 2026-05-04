package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PropertiesListInput struct {
	FileTarget
	Name   string `json:"name,omitempty" jsonschema:"filter to a specific property name"`
	Total  bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Counts bool   `json:"counts,omitempty" jsonschema:"include occurrence counts per property"`
	Active bool   `json:"active,omitempty" jsonschema:"limit scope to the active file"`
	Sort   string `json:"sort,omitempty" jsonschema:"sort order, e.g. 'count'"`
	Format string `json:"format,omitempty" jsonschema:"output format: yaml, json, or tsv"`
}

func propertiesListHandler(ctx context.Context, _ *mcp.CallToolRequest, in PropertiesListInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Name != "" {
		params["name"] = in.Name
	}
	if in.Sort != "" {
		params["sort"] = in.Sort
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Counts {
		flags = append(flags, "counts")
	}
	if in.Active {
		flags = append(flags, "active")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "properties", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type PropertyReadInput struct {
	FileTarget
	Name string `json:"name" jsonschema:"property name to read"`
}

func propertyReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in PropertyReadInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["name"] = in.Name
	out, err := exec.Run(ctx, exec.Args{Command: "property:read", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type PropertySetInput struct {
	FileTarget
	Name  string `json:"name" jsonschema:"property name to set"`
	Value string `json:"value" jsonschema:"property value to assign"`
	Type  string `json:"type,omitempty" jsonschema:"property type: text, list, number, checkbox, date, datetime"`
}

func propertySetHandler(ctx context.Context, _ *mcp.CallToolRequest, in PropertySetInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["name"] = in.Name
	params["value"] = in.Value
	if in.Type != "" {
		params["type"] = in.Type
	}
	out, err := exec.Run(ctx, exec.Args{Command: "property:set", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type PropertyRemoveInput struct {
	FileTarget
	Name string `json:"name" jsonschema:"property name to remove"`
}

func propertyRemoveHandler(ctx context.Context, _ *mcp.CallToolRequest, in PropertyRemoveInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["name"] = in.Name
	out, err := exec.Run(ctx, exec.Args{Command: "property:remove", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterProperties(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_properties",
		Description: "List frontmatter properties in the vault. Optional file/path scope, name filter, counts/total flags, sort, and format (yaml|json|tsv).",
	}, propertiesListHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_property_read",
		Description: "Read the value of a frontmatter property from a note. Provide name; optionally file or path.",
	}, propertyReadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_property_set",
		Description: "Set a frontmatter property on a note. Required name and value; optional type (text|list|number|checkbox|date|datetime).",
	}, propertySetHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_property_remove",
		Description: "Remove a frontmatter property from a note. Provide name; optionally file or path.",
	}, propertyRemoveHandler)
}
