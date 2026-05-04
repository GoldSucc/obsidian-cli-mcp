# Development

## Build & install

```bash
go build ./...           # verify
go install .             # → $GOBIN/obsidian-cli-mcp
```

## Register with Claude

Claude Code (user scope, all projects):

```bash
claude mcp add obsidian-cli -s user $(go env GOPATH)/bin/obsidian-cli-mcp
```

Claude Desktop — edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian-cli": {
      "command": "/Users/<you>/go/bin/obsidian-cli-mcp"
    }
  }
}
```

Optional env: `OBSIDIAN_DEFAULT_VAULT=<name>` to bake in a default vault target.

## Adding a new typed tool

1. Pick its group file in `internal/tools/`. If none fits, add a new file with `Register<Group>(s *mcp.Server)`.
2. Define the input struct with `json` and `jsonschema` tags. Embed `FileTarget` if the command targets a file.
3. Write the handler. Build `Params` map and `Flags` slice. Call `exec.Run`.
4. Append `mcp.AddTool(...)` inside the group's `Register` function.
5. If a new group: add `tools.Register<Group>(server)` to `main.go`.
6. `go build ./...` and `go install .`.
7. Restart Claude Code / Claude Desktop to pick up the new tool schema.

## Gotchas

> [!warning] jsonschema tag descriptions: never lead with `WORD=...`
> The Go SDK's `jsonschema-go` parser interprets a leading `WORD=` as a directive and panics at server startup with `tag must not begin with 'WORD='`. Symptom: MCP client shows "Failed to connect" with no useful message; binary panics on stdin EOF.
>
> **Fix**: rephrase. Bad: `"key=value pairs of params"`. Good: `"params as name/value pairs"`.

> [!warning] Stderr noise from obsidian CLI
> Every `obsidian` invocation writes preamble lines to stderr:
> ```
> 2026-05-04 13:56:24 Loading updated app package /Users/.../obsidian-1.12.7.asar
> Your Obsidian installer is out of date. Please download the latest installer ...
> ```
> `internal/exec/exec.go` strips these before surfacing errors. If you change the binary's stderr format (Obsidian update), update `preamblePatterns` in that file.

> [!warning] Multiline content
> The CLI consumes literal `\n` / `\t` escape sequences in `content=<value>`, not raw newlines. JSON inputs from MCP clients contain real newline chars. `exec.EncodeMultiline` translates — every tool that accepts `content` must call it before writing into the params map.

## Smoke tests

After build + install, restart Claude Code, then:

```
mcp__obsidian-cli__obsidian_version           → "1.12.7 (installer 1.12.7)"
mcp__obsidian-cli__obsidian_vaults verbose=true → tab-separated list
mcp__obsidian-cli__obsidian_search query=<token> → matching file paths
```

## Excluded surface

Intentionally not typed (use `obsidian_run` if needed):

- `plugin*`, `theme*`, `snippet*`, `themes`, `plugins*` — community/plugin/theme management
- `dev:cdp`, `dev:console`, `dev:css`, `dev:debug`, `dev:dom`, `dev:errors`, `dev:mobile`, `dev:screenshot` — devtools
- `devtools`, `eval` — JS execution + Electron internals

Reasoning: these are dev-flow rather than knowledge-management; omitting keeps the LLM-visible tool list focused. Still reachable for one-offs.
