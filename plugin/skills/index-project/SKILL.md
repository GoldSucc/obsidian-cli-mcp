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

## Bootstrap workflow (no index exists)

1. **Detect project name** — `basename($PWD)`.
2. **Scan top-level**: `README.md`, `CLAUDE.md`, `docs/`, `package.json` / `go.mod` / `Cargo.toml` / `pyproject.toml`, `Dockerfile`, CI configs.
3. **Map domains** — open `src/`, `internal/`, `lib/`, `tests/`. Identify subsystems by directory naming + import patterns.
4. **Write `index.md`** first — short overview, link table to all topic pages, last-updated date, project tags.
5. **Write topic pages** — one per identified domain. Use the page template below.
6. **Write `_index.base`** — drop the project base file (template below). Auto-scopes to the project folder via `this.file.folder`. Gives the user a dashboard view of every topic page with last-updated, staleness, group-by-topic.
7. **Cross-link generously** — every wikilink in a page should resolve.
8. **Confirm with user** — show the index map before declaring done.

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

## Files involved

- `path/to/file.ext:42` — what this part does
- `internal/foo/bar.go` — symbol-level pointers welcome

## Gotchas

- Surprises specific to this topic. Cross-link to [[gotchas]] for repo-wide rules.

## Related

[[index]] · [[architecture]] · [[conventions]] · [[<other-topic>]]
```

## Page rules

- **Append over rewrite.** Knowledge accumulates. Don't lose history. Use `obsidian_append`, not `obsidian_create overwrite=true`.
- **Wikilinks resolve by filename** — keep names unique within `semantic-index/<project>/`.
- **Ground every claim in code.** Every assertion needs a `path:line` reference, otherwise it'll rot.
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
