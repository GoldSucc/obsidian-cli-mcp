# obsidian-cli-mcp

MCP server wrapping the official `obsidian` CLI. Lets agents read, write, search, and manage notes in a running Obsidian app.

## Map

- [[architecture]] — layout, exec helper, registration pattern
- [[tool-reference]] — all 58 typed tools + escape hatch, grouped
- [[development]] — adding tools, gotchas, build/install

## Why CLI not REST

`mcp-obsidian` (REST plugin) needs a running Local REST API plugin + API key. Surface ≈ 13 tools. Limited.

`obsidian` CLI ships with the app. Surface ≈ 80 commands. Includes plugin reload, eval, dev console, screenshot, history, sync, bases — none of which the REST plugin exposes.

> [!info] Trade-off
> CLI requires the Obsidian desktop app to be open. Headless workflows still need REST. For dev + personal use, CLI wins.

## Tool prefix

All tools surface as `mcp__obsidian-cli__obsidian_<command>`. Colons in CLI command names → underscores (`daily:read` → `obsidian_daily_read`).
