---
name: index-project
description: Use at the start of any session in a code project repository — read the project's semantic index from `semantic-index/<project>/` in the Obsidian vault before responding to the user's first task. Also use when bootstrapping a new project (no index yet), learning a new pattern or gotcha, recording an architectural decision, before non-trivial changes (read related topics first), or after completing work (append findings). Maintains a continuously growing knowledge base of linked notes plus an `_index.base` dashboard. TRIGGER on: session start in a project, first task touching an unfamiliar subsystem, "remember X" / "important: X" / "gotcha:" from user, completing a non-trivial change, user asks "what's in this project?" or "how does X work here?".
---

# index-project

Claude-managed knowledge base for the **current code project**, persisted in the Obsidian vault. Survives sessions. Compounds over time. All persistent knowledge about a project — architecture, conventions, gotchas, decisions — flows through here.

> [!info] Why this exists
> Memory files (`~/.claude/.../memory/*.md`) are bag-of-strings with weak semantics and no graph. The Obsidian vault is structured: wikilinks form a graph, frontmatter encodes typed metadata, search is fuzzy-aware, the user can read and edit alongside Claude. This skill turns the vault into Claude's project memory.

## Vault layout

```
<vault>/semantic-index/<project>/
├── _index.base       dashboard — queryable view of all topic pages
├── index.md          overview + map (linked from session-start hook)
├── architecture.md   layout, build, deploy, dependencies
├── domain.md         business concepts, entities, vocabulary
├── conventions.md    coding patterns, style decisions, naming
├── gotchas.md        surprises, workarounds, bug fixes, "do NOT" rules
├── decisions.md      ADRs — *why* things are the way they are
├── workflows.md      common dev workflows, commands, scripts
└── <feature>.md      one page per major feature/subsystem (auth, routing, billing, ...)
```

Project name = `basename($PWD)`. Stable across sessions.

`_index.base` turns the project folder into a project-specific database. See **Project base** section below.

## When to invoke

| Trigger | Action |
|---|---|
| Session start | Hook auto-injects `index.md`. Skim it. If user's task is non-trivial, read relevant topic pages too. |
| First time on a project | Bootstrap — full sweep, write all topic pages. |
| User asks "what's in this project?" | Read `index.md`, summarize, point at relevant pages. |
| Before implementing a feature | Read related topic pages first (e.g. `auth.md` before touching auth). |
| After learning something | Append to relevant page using `obsidian_append`. |
| User says "remember X" / "important: X" | Append to `gotchas.md` or `decisions.md`. |
| Architectural change | Update `architecture.md` and any affected feature pages. |

## Bootstrap is two steps — micro first, topics second

| Step | Output | Who runs it | Audience |
|---|---|---|---|
| **1. Microindexing** | `semantic-index/<project>/index/<kind>/<name>.md` — one tiny note per **indexable unit** (source file, ABAP class/CDS/BDef via MCP, config key, env var, HTTP endpoint, DB table, IaC module — anything fetchable) with `#topic/*` tags + anchors | `microindexer` agent (Haiku) | Searching: tag/property queries, "semantic LSP" |
| **2. Topic indexing** (this skill) | `semantic-index/<project>/{index.md, _index.base, <topic>.md, ...}` — prose pages explaining the project's domains, **with explicit wikilinks to the microindex notes that implement each topic** | Parent (Opus) inline OR `documenter` agent (Sonnet) for heavy bootstrap | Reading: humans + Claude reasoning about the project |

Why this order: the microindex is the ground truth — every fetchable unit with its tags. Topic pages synthesize across units. Easier to write the high-level synthesis when the low-level inventory already exists.

**Bidirectional linking**:
- Topic pages contain wikilinks to specific microindex notes (`[[index/<kind>/<name>]]`) under "Implements".
- Microindex notes do NOT manually link back — Obsidian's auto-backlink graph handles the reverse direction. Use `obsidian_backlinks` from a microindex note to find all topic pages that reference it.

## Step 1 — Microindexing (delegate to `microindexer`)

Dispatch the `microindexer` agent first, before writing any topic pages. The agent walks every indexable unit and writes one tiny note per unit under `semantic-index/<project>/index/<kind>/<name>.md`. Each note is frontmatter + 1-line summary + anchors with inline `#topic/*` tags.

Units include: source files (any language), ABAP repository objects (via the SAP ADT MCP), config keys (YAML/TOML/JSON/env), HTTP/RPC endpoints, DB schema objects, IaC modules (Terraform, K8s, Helm), feature flags, env vars. The agent picks the right source per unit (filesystem, MCP, DB introspection, …).

How to dispatch:

```
Agent({
  description: "Microindex <project>",
  subagent_type: "microindexer",
  prompt: "Build microindex for project <project> at <abs path>. Sources: <filesystem dirs and/or ABAP packages and/or config files and/or DB schemas — whatever applies>. Modules: <list-or-flat>. Reuse #topic/* tags from the vault-wide taxonomy. Themes available: <list of docs/<theme>/ to link to>. (Topic pages don't exist yet — link `themes:` to docs/<theme>/, leave topic backlinks for step 2 to add forward.)"
})
```

`microindexer` runs on **Haiku** — cheap mass scanning. Don't pass `model: opus` or `model: sonnet`.

The microindex layout, note template, anchor format, granularity rules, and base file template all live inside the `microindexer` agent file.

## Step 2 — Topic indexing workflow (using the microindex as input)

1. **Detect project name** — `basename($PWD)`.
2. **Read the microindex first** — `obsidian_files_list folder=semantic-index/<project>/index` and sample notes to learn what units exist, by `kind`. The microindex is your map of the project; topics synthesize across it.
3. **Scan top-level for context**: `README.md`, `CLAUDE.md`, `docs/`, `package.json` / `go.mod` / `Cargo.toml` / `pyproject.toml`, `Dockerfile`, CI configs.
4. **Identify topics from the microindex**: cluster microindex notes by tags and themes. Each cluster ≈ one topic page (`authentication`, `data-model`, `deployment`, …).
5. **Write `index.md`** — short overview, link table to all topic pages, last-updated date, project tags.
6. **Write topic pages** — one per identified domain. **Each topic page MUST include an `## Implements` section with wikilinks to the relevant microindex notes** (`[[index/<kind>/<name>]]`). Use the page template below.
7. **Write `_index.base`** — drop the project base file (template below). Auto-scopes via `this.file.folder`.
8. **Cross-link generously** — every wikilink in a page should resolve. Topic pages link to microindex notes via wikilinks; microindex backlinks appear automatically in Obsidian.
9. **Confirm with user** — show the index map before declaring done.

When the topic-page work is heavy (≥5 pages, lots of microindex notes to digest), delegate to the `documenter` agent. The agent's prompt should include: "First, read existing microindex at `semantic-index/<project>/index/`. Cluster microindex notes by tag/theme to identify topics. Each topic page must reference the microindex notes that implement it via wikilinks under `## Implements`."

## Re-running

- **After major refactor**: re-run `microindexer` to refresh anchors and tags. Existing notes get `last-indexed` bumped; new units get new notes; deleted units leave stale notes (surface via the base's "Stale (>30d)" view).
- **After a new feature lands**: parent inline can add a microindex note for the new unit AND extend the relevant topic page with a wikilink to it. Don't redispatch `microindexer` for one or two units.
- **After major topic restructuring**: dispatch `documenter` to rewrite affected topic pages, re-reading the microindex.

For small changes (one new unit, one renamed class) the parent does the update inline — no need to dispatch agents for ≤5 writes.

```
mcp__obsidian-cli__obsidian_create  path="semantic-index/<project>/index.md"  content="..."
mcp__obsidian-cli__obsidian_create  path="semantic-index/<project>/architecture.md"  content="..."
... etc
```

## Update workflow (index exists)

1. **Read `index.md`** to know what topics already exist (the SessionStart hook auto-loads this — usually you've already seen it).
2. **Read topic pages** relevant to the current task before touching code.
3. **After completing a task**: identify what changed in your understanding. Append to relevant page with `obsidian_append`.
4. **If a new domain emerges**: create a new topic page (`obsidian_create`), add a row in `index.md` (`obsidian_append`), bump the index's "last updated" line.
5. **Date every change** — set `last-updated` in frontmatter via `obsidian_property_set name="last-updated" value="<YYYY-MM-DD>" type="date"`.

## Page template

```markdown
---
project: <project>
topic: <topic>
last-updated: 2026-05-04
tags: [project/<project>, topic/<topic>]
---

# <Topic>

## Overview

<2-3 sentence summary — what this subsystem is, why it exists>

## Details

- short fragments, headings
- group related points
- prefer tables for structured data

## Implements

The microindex notes (in `index/<kind>/`) that implement this topic. Forward links here; Obsidian shows the reverse via auto-backlinks.

- [[index/file/auth-middleware-go]] — JWT middleware
- [[index/class/ZCL_AUTH_HANDLER]] — SAML SSO handler
- [[index/cds/Z_USER_VIEW]] — User attribute exposure
- [[index/config/jwt-secret]] — Token signing key
- [[index/endpoint/POST-api-auth-login]] — Login endpoint

## References

A future LLM reading this page must be able to **fetch the referenced thing without searching**. Every behavioral claim in Details gets at least one entry here. The microindex notes linked above already carry the full `ref` — restate only when the topic page makes a claim about a specific line/method that the microindex doesn't already capture, or when the reference is outside the microindex (external URL, command, etc.). Reference taxonomy:

- `internal/auth/middleware.go:42-58` — file with line range
- `class:ZCL_AUTH_HANDLER` (S/4HANA, package `ZAUTH`)
- `cds:Z_USER_VIEW`
- `function:Z_USER_FETCH` (FuGr `ZUSR`)
- `config:auth.jwt.secret` in `.env`
- `env:DATABASE_URL` defined in `docker-compose.yml`
- `table:ZAUTH_LOG` (custom DB table)
- `endpoint:POST /api/auth/login` (handled in `internal/auth/handler.go:18`)
- `url:https://...` (RFC, blog, doc)
- `command:make build`

## Gotchas

- Surprises specific to this topic. Cross-link to [[gotchas]] for repo-wide rules.

## Related

[[index]] · [[architecture]] · [[conventions]] · [[<other-topic>]]
```

## Reference taxonomy

A "reference" is anything an LLM can later fetch and inspect. Pick the prefix that matches the kind:

| Prefix | Pattern | Fetch with |
|---|---|---|
| (no prefix, file path) | `path/to/file.ext:42` or `path/to/file.ext:42-58` or `path/to/file.ext` | `Read` (with line range when applicable) |
| `class:` | `class:ZCL_FOO` | `mcp__plugin_vsp_sap-adt__GetSource` (ABAP) / `Read` for non-ABAP |
| `interface:` | `interface:ZIF_BAR` | language-appropriate symbol fetch |
| `function:` | `function:Z_FOO` (FuGr `ZBAR`) | `mcp__plugin_vsp_sap-adt__GetSource` |
| `cds:` | `cds:Z_VIEW` | `mcp__plugin_vsp_sap-adt__GetSource` |
| `bdef:` | `bdef:Z_BO_DEF` | `mcp__plugin_vsp_sap-adt__GetSource` |
| `srvd:` / `srvb:` | `srvd:Z_SERVICE_DEF` | `mcp__plugin_vsp_sap-adt__GetServiceMetadata` |
| `dtel:` / `doma:` / `tabl:` / `strucutre:` | `tabl:ZAUTH_LOG` | `mcp__plugin_vsp_sap-adt__GetTable` / `GetSource` |
| `config:` | `config:section.key` in `<file>` | `Read` the config file, locate key |
| `env:` | `env:VAR_NAME` defined in `<file>` | `Read` the file (compose, .env, dockerfile) |
| `endpoint:` | `endpoint:METHOD /path` (handled in `<file>:<line>`) | `Read` the handler file |
| `table:` | `table:NAME` (in DB X) | DB introspection; for ABAP `GetTable` |
| `package:` | `package:ZFOO` | `mcp__plugin_vsp_sap-adt__GetPackage` |
| `transport:` | `transport:DEVK900123` | `mcp__plugin_vsp_sap-adt__GetTransport` |
| `command:` | `command:make build` | `Bash` |
| `url:` | `url:https://...` | `WebFetch` / `obsidian:defuddle` |

Always include the **smallest fetchable unit**:
- File + line range > file alone > directory.
- ABAP object name > package alone.
- Specific config key > "see config file".

If a reference doesn't fit a prefix, write `<your-prefix>:<id>` and add a parenthetical hint of how to fetch it. The LLM will figure it out.

## Page rules

- **Append over rewrite.** Knowledge accumulates. Don't lose history. Use `obsidian_append`, not `obsidian_create overwrite=true`.
- **Wikilinks resolve by filename** — keep names unique within `semantic-index/<project>/`.
- **Ground every claim in a reference.** Every assertion has at least one entry under `## References` or it doesn't go in the page. References must be fetchable (line range, object name, config key, env var, endpoint) — not vague pointers like "in the auth module".
- **Caveman style**: short fragments, headings, tags, embeds. Match the user's vault style.
- **Frontmatter fields** are typed via `obsidian_property_set`. `last-updated` is a `date`. `tags` is a `list`.
- **Tag everything `#project/<name>`** — makes the project's index queryable from a Bases view across the vault.

## index.md template

```markdown
---
project: <project>
last-updated: 2026-05-04
tags: [project/<project>, semantic-index]
---

# <Project> — Semantic Index

> [!info] What this is
> Claude-managed knowledge base for the `<project>` repo. Updated continuously across sessions. Read this first; pull topic pages as needed.

## Map

| Topic | Page |
|---|---|
| Architecture | [[architecture]] |
| Domain | [[domain]] |
| Conventions | [[conventions]] |
| Gotchas | [[gotchas]] |
| Decisions | [[decisions]] |
| Workflows | [[workflows]] |
| ... | ... |

## Project pulse

- Repo path: `<absolute path>`
- Primary languages: <Go, TypeScript, ...>
- Build: `<command>`
- Test: `<command>`
- Run: `<command>`

## Dashboard

![[_index.base]]

## Recent changes

- 2026-05-04: bootstrapped index
- ... (append on every meaningful update)
```

## Project base (`_index.base`)

The base turns the project's index folder into a queryable dashboard: every topic page becomes a row, with frontmatter as columns, formulas for staleness, grouped views.

**Filename**: `_index.base` (underscore prefix sorts it next to `index.md` in the file pane).

**Scope**: filters via `file.inFolder(this.file.folder)` so the base auto-scopes to its parent folder. No project name hardcoded — copy-paste-safe across projects.

**Loading the obsidian-bases skill** when authoring the base is helpful for syntax (filters, formulas, view types). The skill is at `obsidian:obsidian-bases`.

### Template

```yaml
filters:
  and:
    - file.inFolder(this.file.folder)
    - 'file.ext == "md"'

formulas:
  days_since_update: 'if(last-updated, (today() - date(last-updated)).days, "")'
  is_stale: 'if(last-updated, (today() - date(last-updated)).days > 30, true)'
  page_type: 'if(file.basename == "index", "overview", topic)'

properties:
  topic:
    displayName: "Topic"
  last-updated:
    displayName: "Updated"
  formula.days_since_update:
    displayName: "Days Old"
  formula.page_type:
    displayName: "Type"

views:
  - type: table
    name: "All topics"
    order:
      - file.name
      - topic
      - last-updated
      - formula.days_since_update
    groupBy:
      property: topic
      direction: ASC

  - type: table
    name: "Stale (>30d)"
    filters:
      and:
        - 'formula.is_stale == true'
        - 'file.basename != "index"'
    order:
      - file.name
      - last-updated
      - formula.days_since_update

  - type: list
    name: "Recently updated"
    limit: 10
    order:
      - file.name
      - last-updated
```

### Embedding

You can also embed the base inline in `index.md`:

```markdown
## Dashboard

![[_index.base]]
```

Renders the table directly in the index page.

### When to write the base

- During bootstrap, after topic pages exist (so the base has something to show).
- When the index gains a new structured frontmatter field worth surfacing as a column — add a row to `properties:` and update view `order:`.
- Don't overload it with topic-specific views — keep it generic. Topic-specific bases belong in their own files (e.g. `gotchas.base` for a more detailed gotcha browser).

## Tools

All operations use the `obsidian-cli` MCP. Common calls:

| Goal | Tool |
|---|---|
| Read full page | `mcp__obsidian-cli__obsidian_read path="semantic-index/<project>/<topic>.md"` |
| Append finding | `mcp__obsidian-cli__obsidian_append path="..." content="..."` |
| Create new topic | `mcp__obsidian-cli__obsidian_create path="..." content="..."` |
| Search within index | `mcp__obsidian-cli__obsidian_search query="..." path="semantic-index/<project>"` |
| Read frontmatter prop | `mcp__obsidian-cli__obsidian_property_read name="last-updated" path="..."` |
| Set frontmatter prop | `mcp__obsidian-cli__obsidian_property_set name="last-updated" value="2026-05-04" type="date" path="..."` |
| List existing topics | `mcp__obsidian-cli__obsidian_files_list folder="semantic-index/<project>"` |
| Outline of a page | `mcp__obsidian-cli__obsidian_outline path="..." format=md` |
| Create the base | `mcp__obsidian-cli__obsidian_create path="semantic-index/<project>/_index.base" content="<yaml>"` |
| Query the base | `mcp__obsidian-cli__obsidian_base_query path="semantic-index/<project>/_index.base" view="All topics" format=json` |

## Delegate to the `documenter` subagent for heavy work

This skill works two ways: parent agent executes inline for small ops, OR delegates to the `documenter` Sonnet subagent for heavier writes.

**Inline (parent does directly)**:
- Reading the index, 1-2 topic pages
- Appending a single gotcha or decision
- Bumping `last-updated` on one page
- Quick search across the index

**Delegate to `documenter`**:
- **Bootstrap** — building the index for the first time (writes 6+ pages + `_index.base`)
- **Mass extend** — updating many pages after a structural change (architecture migration, dependency overhaul)
- Anything that'd be >5 MCP write tool calls

How to delegate:

```
Agent({
  description: "Bootstrap semantic-index/<project>/",
  subagent_type: "documenter",
  prompt: "Bootstrap semantic-index/<project>/ from the repo at <abs path>. Cover architecture, domain, conventions, gotchas, decisions, workflows. Ground every claim in path:line. Use the index-project skill spec. Sources: the repo itself."
})
```

The agent runs on Sonnet (cheap + fast for the structured-write workload), invokes this skill itself, executes the bootstrap, returns a structured report. Don't pass `model: opus` when invoking — `documenter` is sonnet by design.

## Anti-patterns

- ❌ Storing project knowledge in `~/.claude/.../memory/*.md`. Wrong store. Migrate to vault.
- ❌ Generic facts in `gotchas.md` ("always use type hints"). Belongs in user-level memory, not a per-project index.
- ❌ Page named `notes.md` or `misc.md`. Topics must be specific. If it doesn't fit a topic, it shouldn't be in the index.
- ❌ Wikilinks to non-existent pages. Resolve them or remove them.
- ❌ Re-running bootstrap when index exists. Update incrementally instead.

## Boundaries

- This skill manages **code projects** under `semantic-index/`. It does **not** touch `Projects/<Customer>/` (human-curated SAP/customer notes — different namespace).
- The index is **descriptive, not prescriptive**. It records what *is*, including known gotchas. Use `decisions.md` for *why* things are.
- Don't index secrets, credentials, or anything `.gitignore`d.
