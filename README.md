# obsidian-cli-mcp

MCP server + Claude Code plugin that wraps the official `obsidian` CLI and turns the Obsidian vault into Claude's persistent project memory.

## What's in the box

- **MCP server** with 58 typed tools covering reads, writes, search, daily notes, properties, tasks, history, sync, bookmarks, bases, workspace.
- **`obsidian_run`** escape hatch for any CLI command not yet typed (plugin/theme/dev/eval).
- **`index-project` skill** — Claude-maintained semantic index of code projects under `semantic-index/<project>/` in your vault. Linked notes, frontmatter, wikilinks. Survives sessions, compounds over time.
- **`document-theme` skill** — general topic knowledge base under `docs/<theme>/` with `_index.base` dashboards. Subject-noun themes, source-cited (Context7, defuddle, web), reusable across projects.
- **`documenter` subagent** (Sonnet) — specialized for mass-indexing and theme documentation. Parent Opus delegates focused doc tasks; documenter executes against the skills.
- **`microindexer` subagent** (Haiku) — second-step indexer that walks every **indexable unit** in a project — source files, ABAP repository objects (via SAP ADT MCP), config keys, env vars, HTTP endpoints, DB tables, IaC modules — and creates one tiny semantic-only note per unit under `semantic-index/<project>/index/<kind>/<name>.md`. Each note is frontmatter + 1-line summary + anchors with inline `#topic/*` tags. Designed for retrieval, not reading — a queryable "semantic LSP" so Claude can ask "what implements login?" via `obsidian_tag` and get the full list of relevant units regardless of where they live.
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
mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_version           → "1.12.7 (installer 1.12.7)"
mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_vaults verbose=true
```

If the project doesn't have a semantic index yet, ask Claude to invoke the `index-project` skill or dispatch the `documenter` agent (see below).

## Knowledge workflow

The plugin turns your Obsidian vault into Claude's persistent memory. Two stores, two skills, one specialized subagent.

### Three stores in your vault

| Path | Store | Audience |
|---|---|---|
| `semantic-index/<project>/<topic>.md` | Per-code-project topic pages. Prose explaining the project's domains. | Claude reasoning + humans |
| `semantic-index/<project>/index/<kind>/<name>.md` | Per-code-object microindex. Frontmatter + line-range anchors + `#topic/*` tags. Semantic LSP. | Tag/property queries — "what implements X?" |
| `docs/<theme>/` | Per-topic general knowledge. Curated, source-cited deep dives (OAuth, Kubernetes, regex, …). | Humans + Claude |

A top-level `docs/_index.base` dashboard catalogues every theme. A per-project `semantic-index/<project>/_index.base` dashboards the topic pages. A per-project `semantic-index/<project>/index/_index.base` pivots the microindex by `kind`, `module`, or `theme`.

### Two skills, both auto-loaded

- **`obsidian-cli-mcp:index-project`** — bootstrap & maintain `semantic-index/<project>/`. Triggered at session start (via hook) and whenever a gotcha, decision, or new pattern lands.
- **`obsidian-cli-mcp:document-theme`** — bootstrap & extend `docs/<theme>/` for general subjects. Triggered when you say "document X", "research Y", "build a knowledge base on Z".

Both skills define page templates, frontmatter contracts, and `_index.base` YAML so the structure stays consistent.

### `documenter` — Sonnet subagent for mass writes

Heavy documentation work (bootstrapping a fresh index/theme, mass-extending after a refactor, research-heavy theme creation) shouldn't burn Opus tokens. The plugin ships a dedicated `documenter` subagent pinned to **Sonnet** for exactly this workload.

**Why a separate agent?**

- Cheaper + faster for structured-write workloads
- Isolated context — the parent's reasoning chain stays clean
- Hard-coded to additive-only MCP tools (no delete / move / rename / property_remove / `obsidian_run`) so it can never destroy notes

**When the parent delegates to it:**

| Workload | Path |
|---|---|
| ≤5 MCP write calls (single append, property bump, quick gotcha) | Inline (parent does it) |
| Bootstrap a whole `semantic-index/<project>/` | Delegate to `documenter` |
| Bootstrap a fresh `docs/<theme>/` with web research | Delegate to `documenter` |
| Mass extend after architectural change | Delegate to `documenter` |

**How to dispatch:**

```
Agent({
  description: "Bootstrap obsidian-cli-mcp project index",
  subagent_type: "documenter",
  prompt: "Bootstrap semantic-index/obsidian-cli-mcp/ from the repo at /Users/dz434/obsidian-cli-mcp. Cover architecture (Go MCP server + plugin layout), conventions, gotchas (jsonschema WORD= bug, stderr preamble), decisions, workflows. Ground every claim in path:line. Use the index-project skill spec."
})
```

```
Agent({
  description: "Bootstrap docs/oauth/",
  subagent_type: "documenter",
  prompt: "Bootstrap docs/oauth/ in the vault. Subject: OAuth 2.1 + OIDC. Subpages: overview, concepts, howto, examples, gotchas, references. Sources: the OAuth 2.1 draft RFC, OIDC core spec, Auth0 / Okta docs via Context7. Cite every claim. Synthesize, don't paste."
})
```

The agent invokes the relevant skill itself, executes against templates, returns a ≤30-line structured report listing paths created, sources used, gaps left for follow-up.

### `microindexer` — Haiku subagent for the semantic LSP layer

After topic pages exist, dispatch `microindexer` to walk every code object and produce one tiny note per object under `semantic-index/<project>/index/<kind>/<name>.md`. Each note has:

- Frontmatter: `kind`, `name`, `ref`, `path`, `tags`, `themes`, `last-indexed`
- 1-line semantic summary
- **Anchors** — bullet per searchable region with `path:line-range` + `#topic/*` tags + optional `[[topic-page]]` wikilinks

Example anchors inside `semantic-index/techstep/index/file/auth.md`:

```
- `Login` — `internal/auth/auth.go:23-55` — login impl, password check, JWT issue #topic/auth #topic/login #topic/jwt [[../../authentication]]
- `validateCredentials` — `internal/auth/auth.go:60-78` — bcrypt compare #topic/auth #topic/password
- `IssueJWT` — `internal/auth/auth.go:80-100` — sign claims with HS256 #topic/jwt #topic/tokens [[../../authentication]]
```

This makes the microindex queryable Obsidian-natively:

- `obsidian_tag name=topic/login verbose path=semantic-index/<project>` → every code object related to login
- `obsidian_search query="topic/jwt" path=semantic-index/<project>/index` → files with JWT-tagged anchors
- Bases dashboard at `semantic-index/<project>/index/_index.base` → pivot by `kind`, by `module`, by `theme`, by recency

**How to dispatch:**

```
Agent({
  description: "Microindex techstep code objects",
  subagent_type: "microindexer",
  prompt: "Build microindex for project techstep at <abs path>. Scope: src/ + abap/. Modules: web (TypeScript), abap (S/4HANA). Reuse existing #topic/* and #theme/* tags from semantic-index/techstep/. Cross-link to topic pages and docs/sap-abap/, docs/cap/."
})
```

The agent runs on Haiku (cheap mass scanning), produces hundreds of small notes, returns a count-by-kind report.

> [!info] Don't override agent models
> Both agents pin a model in their definition: `documenter` is Sonnet, `microindexer` is Haiku. Don't pass `model: opus` when invoking — that defeats the cost split. Pass only `subagent_type` and `prompt`.

### Triggering paths summary

```
You type:           "use documenter to ..."           → direct dispatch
You type:           "document <theme>"                 → parent reads document-theme skill
                                                          → if heavy, parent dispatches documenter
                                                          → if light, parent does inline
Session starts:     SessionStart hook injects index → parent applies index-project skill as needed
Gotcha learned:     parent's own judgment              → append via index-project skill (inline)
Big refactor:       parent's own judgment              → dispatch documenter for mass extend
```

## Layout

```
.claude-plugin/
  marketplace.json              Claude marketplace metadata (single-plugin)
plugin/                         the plugin itself
  .claude-plugin/
    plugin.json                 plugin manifest
  .mcp.json                     registers the obsidian-cli-mcp binary as an MCP server
  hooks/
    hooks.json                  SessionStart hook config
    session_start.py            reads semantic-index/<project>/index.md from vault, injects as context
  skills/
    index-project/SKILL.md      project memory skill (semantic-index/<project>/)
    document-theme/SKILL.md     topic knowledge skill (docs/<theme>/)
  agents/
    documenter.md               Sonnet subagent for heavy doc work
docs/                           repo documentation (architecture, tool reference, dev notes)
main.go, internal/, go.mod      MCP server source
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
