# Architecture

## Layout

```
obsidian-cli-mcp/
├── go.mod                       module github.com/GoldSucc/obsidian-cli-mcp
├── main.go                      server boot, transport, group registration
├── internal/
│   ├── exec/exec.go             shell out to `obsidian`, strip stderr preamble
│   └── tools/
│       ├── generic.go           obsidian_run — escape hatch
│       ├── files.go             read, create, append, prepend, move, rename, delete, file_info, files_list, wordcount
│       ├── folders.go           folder, folders
│       ├── daily.go             daily, daily_read, daily_append, daily_prepend, daily_path
│       ├── search.go            search, search_context, search_open
│       ├── tasks.go             task, tasks
│       ├── properties.go        properties, property_read, property_set, property_remove
│       ├── tags_links.go        tag, tags, aliases, links, backlinks, unresolved, orphans, deadends, outline
│       ├── random.go            random, random_read
│       ├── templates.go         templates, template_read, template_insert
│       ├── bases.go             bases, base_views, base_query, base_create
│       ├── bookmarks.go         bookmark, bookmarks
│       ├── history.go           history, history_list, history_read, history_restore, history_open, diff
│       ├── sync.go              sync, sync_status, sync_deleted, sync_history, sync_read, sync_restore, sync_open
│       ├── vault.go             vault_info, vaults, reload, restart, version
│       ├── workspace.go         workspace, tabs, tab_open, recents
│       └── commands.go          command, commands, hotkey, hotkeys
└── docs/                        you are here
```

## Exec helper (`internal/exec/exec.go`)

Every tool delegates to one function:

```go
exec.Run(ctx, exec.Args{
    Command: "<obsidian subcommand>",
    Vault:   "<vault name or empty>",
    Params:  map[string]string{...},  // key=value args
    Flags:   []string{...},            // bare flag args
})
```

Responsibilities:
- Build argv: `[obsidian, vault=<v>?, <command>, k=v..., flag1, flag2]`
- Sort param keys → deterministic command line (helps reproduce bugs)
- Capture stdout / stderr separately
- Strip stderr preamble (Obsidian writes `Loading updated app package...` and `Your Obsidian installer is out of date...` on every invocation — these are not errors)
- On non-zero exit, wrap stripped stderr as Go error
- Honour `OBSIDIAN_DEFAULT_VAULT` env var when `Vault` is empty

There's also `exec.EncodeMultiline(s)` — converts real `\n` / `\t` chars in JSON input to the literal escape sequences the Obsidian CLI expects.

## Tool group pattern

Each `internal/tools/<group>.go` exposes one `Register<Group>(s *mcp.Server)` function. `main.go` wires them up centrally:

```go
tools.RegisterGeneric(server)
tools.RegisterFiles(server)
tools.RegisterFolders(server)
... etc
```

Per-tool shape:

```go
type AppendInput struct {
    FileTarget                              // embedded — vault, file, path
    Content string `json:"content" jsonschema:"text to append; real newlines accepted"`
    Inline  bool   `json:"inline,omitempty" jsonschema:"append without leading newline"`
}

func appendHandler(ctx context.Context, _ *mcp.CallToolRequest, in AppendInput) (*mcp.CallToolResult, TextOutput, error) {
    params := in.params()
    params["content"] = exec.EncodeMultiline(in.Content)
    flags := []string{}
    if in.Inline {
        flags = append(flags, "inline")
    }
    out, err := exec.Run(ctx, exec.Args{Command: "append", Vault: in.Vault, Params: params, Flags: flags})
    if err != nil {
        return nil, TextOutput{}, err
    }
    return nil, TextOutput{Content: out}, nil
}
```

Shared types live in `files.go` (the first file written): `FileTarget`, `TextOutput`. Other groups reuse them — no redeclaration.

## Output shape

Every typed tool returns `TextOutput { Content string }` — raw stdout passthrough. The LLM parses if needed. JSON-format CLI output (`format=json`) is returned verbatim as a string.

## Escape hatch

`obsidian_run` is the catch-all:

```json
{
  "command": "plugin:reload",
  "vault": "My Vault",
  "params": {"id": "my-plugin"},
  "flags": ["silent"]
}
```

Use when a CLI command isn't yet typed (plugin/theme/dev/snippet commands intentionally omitted from the typed surface).
