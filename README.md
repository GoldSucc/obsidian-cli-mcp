# obsidian-cli-mcp

MCP server + Claude Code plugin that wraps the official `obsidian` CLI and turns the Obsidian vault into Claude's persistent project memory.

## What's in the box

- **MCP server** with 58 typed tools covering reads, writes, search, daily notes, properties, tasks, history, sync, bookmarks, bases, workspace.
- **`obsidian_run`** escape hatch for any CLI command not yet typed (plugin/theme/dev/eval).
- **`index-project` skill** — Claude-maintained semantic index of code projects under `semantic-index/<project>/` in your vault. Linked notes, frontmatter, wikilinks. Survives sessions, compounds over time.
- **SessionStart hook** — auto-injects the project's `index.md` at session start so Claude opens with full context.

See [`docs/index.md`](docs/index.md) for architecture, full tool reference, and dev notes.

## Install

### 1. Install the MCP binary

Requires Go 1.22+ and Obsidian desktop app.

```bash
go install github.com/GoldSucc/obsidian-cli-mcp@latest
```

This places `obsidian-cli-mcp` in `$(go env GOPATH)/bin`. Make sure that's on your `PATH`.

The binary shells out to `obsidian` (the official CLI bundled with the Obsidian app). Verify:

```bash
which obsidian          # /Applications/Obsidian.app/Contents/MacOS/obsidian on macOS
obsidian version
```

If `obsidian` isn't on PATH, install the latest Obsidian, enable command line tools in general settings and re-launch your shell.

### 2. Add the marketplace + install the plugin (recommended)

```bash
claude /plugin marketplace add github.com/GoldSucc/obsidian-cli-mcp
claude /plugin install obsidian-cli-mcp@obsidian-cli-mcp
```

Or wire it manually in `~/.claude/settings.json`:

```jsonc
{
  "extraKnownMarketplaces": {
    "obsidian-cli-mcp": {
      "source": {
        "source": "git",
        "url": "https://github.com/GoldSucc/obsidian-cli-mcp.git"
      }
    }
  },
  "enabledPlugins": {
    "obsidian-cli-mcp@obsidian-cli-mcp": true
  }
}
```

### 3. (Claude Desktop) wire the MCP

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "obsidian-cli": {
      "command": "obsidian-cli-mcp"
    }
  }
}
```

### 4. (Optional) bake in a default vault

```bash
export OBSIDIAN_DEFAULT_VAULT="My Vault"
```

When `vault=` isn't passed to a tool, this is used. Otherwise the CLI defaults to the most recently focused vault.

## Verify

Restart Claude Code. The SessionStart hook will print a `Project Index — <name>` banner. From there:

```
mcp__obsidian-cli__obsidian_version           → "1.12.7 (installer 1.12.7)"
mcp__obsidian-cli__obsidian_vaults verbose=true
```

If the project doesn't have a semantic index yet, invoke the `index-project` skill to bootstrap one.

## Layout

```
.claude-plugin/
  marketplace.json      Claude marketplace metadata
  plugin.json           plugin metadata
.mcp.json               registers the obsidian-cli-mcp binary as an MCP server
hooks/
  hooks.json            SessionStart hook config
  session_start.py      reads semantic-index/<project>/index.md from vault, injects as context
skills/
  index-project/
    SKILL.md            the indexing skill
docs/
  index.md              repo overview
  architecture.md       Go server architecture
  tool-reference.md     all 58 typed tools + escape hatch
  development.md        adding tools, gotchas
main.go, internal/      MCP server source
```

## Updating

```bash
go install github.com/GoldSucc/obsidian-cli-mcp@latest
```

Restart Claude Code to pick up new tool schemas.

## Why CLI not REST plugin

The `mcp-obsidian` Python server uses Obsidian's Local REST API plugin. Surface ~13 tools.

The official `obsidian` CLI ships with the desktop app and exposes ~80 commands including plugin reload, eval, dev console, screenshot, history, sync, bases — none of which the REST plugin exposes. Caveat: requires Obsidian desktop app to be open.

## License

MIT.
