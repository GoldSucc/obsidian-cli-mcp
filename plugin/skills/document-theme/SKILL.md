---
name: document-theme
description: Build, extend, and maintain a personal knowledge base on a general topic ("theme") — OAuth flows, Kubernetes networking, Rust async, PostgreSQL replication, ABAP CDS, etc. Researches via web, library docs, and user input; produces structured, linked markdown in the Obsidian vault under `docs/<theme>/` with an `_index.base` dashboard. Topic-scoped, NOT project-scoped (project knowledge belongs to `index-project`). TRIGGER on: "document <topic>", "research and write up <X>", "build a knowledge base on <Y>", "explain <Z> and save it for future reference", "compile notes on <subject>", extending an existing `docs/<theme>/`.

---

# document-theme

Builds and grows a **topic-centric** knowledge base in the Obsidian vault. Each theme is a self-contained reference — a curated mental model on a subject — that you can return to and extend across projects, jobs, years.

> [!info] Topic, not project
> `index-project` covers **the current code project** — what's in the repo, where things live, project-specific gotchas. Lives at `semantic-index/<project>/`. Tied to `basename($PWD)`.
>
> `document-theme` covers **a general subject** — OAuth, distributed consensus, regex, Kafka. Lives at `docs/<theme>/`. NOT tied to any project. Reusable across all your work.
>
> If a thing matters only inside one repo → it's a project gotcha, not a theme. If a thing matters everywhere you encounter the subject → theme.

## When to use

| Trigger | Action |
|---|---|
| User says "document `<topic>`" | Bootstrap `docs/<topic>/` if missing, extend if present. |
| User says "research and write up `<X>`" | Use web + library docs + user input to produce a theme. |
| Reading docs / articles / books on a subject | Capture the durable mental model into a theme. |
| Repeated lookups on the same topic | Stop re-researching; write the theme once. |
| Extending an existing `docs/<theme>/` | Read it, identify gaps, add subpages. |

## Vault layout

```
<vault>/docs/
├── README.md            top-level entry — index of every theme
├── _index.base          aggregated dashboard across all themes
└── <theme>/             e.g. oauth, kubernetes-networking, rust-async, postgres-replication
    ├── README.md        theme entry — scope, audience, source bibliography, nav
    ├── _index.base      theme-scoped dashboard
    ├── overview.md      intro + mental model (read-first)
    ├── concepts.md      key terms, definitions, distinctions
    ├── howto.md         practical recipes
    ├── examples.md      worked examples with annotated code/config
    ├── gotchas.md       common pitfalls, footguns, "do NOT" rules
    ├── references.md    canonical sources — URLs, books, RFCs, papers
    └── <subtopic>.md    more pages as the theme deserves
```

Theme name = `kebab-case`, **subject-noun**: `oauth`, `kubernetes-networking`, `cap-cds`, `rust-async`, `regex`, `linear-algebra`. NOT `claude-prompts-for-X` (verb), NOT `q3-research` (sprint).

Two bases:
- `docs/_index.base` — root, aggregates every theme
- `docs/<theme>/_index.base` — single theme

Both use `file.inFolder(this.file.folder)` so they're portable.

## Bootstrap workflow (no `docs/<theme>/` exists)

1. **Confirm theme + scope** with the user. Where's the bound? OAuth in general, or just OAuth 2.1? Kubernetes networking, or just CNI plugins? Don't guess; ask once if ambiguous.
2. **Initialize root if missing** — if `docs/README.md` doesn't exist in the vault, create both `docs/README.md` (root index template) and `docs/_index.base` (root base template) first. Only on the very first theme.
3. **Outline subpages** — list the pages you intend to create and what each will cover. Confirm with user if unsure. Default subpages: `overview`, `concepts`, `howto`, `examples`, `gotchas`, `references`.
4. **Research the subject** — combine sources, prefer authoritative over blog-quality:
   - **Library / framework docs**: `mcp__plugin_context7_context7__resolve-library-id` then `query-docs`. Prefer over WebFetch for SDK questions.
   - **Web articles, RFCs, specs**: `/obsidian:defuddle <url>` to pull clean markdown into the conversation. Cite URL + retrieval date.
   - **General web search**: `WebSearch` for current state, recent changes, gaps Context7 doesn't fill.
   - **User input**: ask for sources the user already trusts. Don't re-research what they've already vetted.
5. **Synthesize, don't dump.** Read multiple sources, distil into the user's mental model. Don't paste extracted blocks; rewrite in your own words and cite the source.
6. **Write theme `README.md`** — `obsidian_create path="docs/<theme>/README.md" content=<readme>`.
7. **Write each subpage** using the page template. Source citations as URLs, books, paper refs (no `path:line` — this isn't code-tied). Wikilinks for cross-page nav. Frontmatter via `obsidian_property_set` for typed fields.
8. **Write theme `_index.base`** — `obsidian_create path="docs/<theme>/_index.base" content=<base-yaml>`.
9. **Update root `docs/README.md`** — add a row in the Themes table: `obsidian_append path="docs/README.md" content="| [[<theme>/README\|<theme>]] | <one-line scope> |"`.
10. **Cross-link to related themes** if any exist (e.g. `docs/oauth/` should link to `docs/openid-connect/` if both exist).
11. **Confirm with user** — show the layout you produced before declaring done.

## Extend workflow (theme exists)

1. **Read theme `README.md`** — `obsidian_read path="docs/<theme>/README.md"`.
2. **Read relevant subpages** that overlap with the new content.
3. **Decide** — extending existing page (`obsidian_append`) or new subpage (`obsidian_create`)? Default to extending unless the new material is structurally distinct.
4. **Bump `last-updated`** — `obsidian_property_set name="last-updated" value="<YYYY-MM-DD>" type="date" path="..."` on every modified page.
5. **Update theme `README.md` navigation** if a new page was added.
6. **Update `references.md`** with the new source(s) used.
7. **Update theme + root base files** only if you added a new frontmatter field worth surfacing as a column.

## Page template

```markdown
---
theme: <theme>
topic: <topic>
last-updated: 2026-05-04
tags: [theme/<theme>, topic/<topic>]
sources: [<url-or-ref-1>, <url-or-ref-2>]
---

# <Topic>

> [!info] Audience
> <one-line: who this page is for, what they should know after reading>

## Overview

<2-3 sentences — the mental model, not the encyclopedia entry>

## Details

<headings, fragments, tables, diagrams as needed>

## Examples

```<lang>
<minimal, runnable, annotated>
```

## Gotchas

- Specific surprises. Cross-link to [[gotchas]] for the theme-wide list.

## Sources

- <URL 1> — what was useful here
- <RFC / book / paper ref> — chapter or section if applicable

## Related

[[README]] · [[overview]] · [[concepts]] · [[<other-topic>]]
```

## Theme `README.md` template

```markdown
---
theme: <theme>
last-updated: 2026-05-04
tags: [theme/<theme>, docs]
---

# <Theme>

> [!info] Scope
> <what this theme covers, and importantly what it does NOT cover>
>
> **Audience**: <who benefits from reading — practitioners, learners, on-call?>

## Map

| Page | Covers |
|---|---|
| [[overview]] | mental model, big picture |
| [[concepts]] | key terms, definitions |
| [[howto]] | practical recipes |
| [[examples]] | worked examples |
| [[gotchas]] | edge cases, footguns |
| [[references]] | canonical sources |

## Dashboard

![[_index.base]]

## Source bibliography

Top sources for this theme (full list lives in [[references]]):

- <Most-cited URL or book>
- <Authoritative RFC / spec>
- ...

## Related themes

- [[../<related-theme>/README]] — how it relates

## Recent changes

- 2026-05-04: theme bootstrapped — covers <X, Y, Z>
- ... (append on every meaningful update)
```

## Theme `_index.base` template

```yaml
filters:
  and:
    - file.inFolder(this.file.folder)
    - 'file.ext == "md"'

formulas:
  days_since_update: 'if(last-updated, (today() - date(last-updated)).days, "")'
  is_stale: 'if(last-updated, (today() - date(last-updated)).days > 180, true)'
  page_type: 'if(file.basename == "README", "overview", topic)'

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
    name: "All pages"
    order:
      - file.name
      - topic
      - last-updated
      - formula.days_since_update
    groupBy:
      property: topic
      direction: ASC

  - type: table
    name: "Stale (>180d)"
    filters:
      and:
        - 'formula.is_stale == true'
        - 'file.basename != "README"'
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

## Root `docs/README.md` template

Created on the very first theme bootstrap if `docs/` is empty.

```markdown
---
last-updated: 2026-05-04
tags: [docs, root-index]
---

# Knowledge base

> [!info] Audience
> Anyone needing durable, structured notes on a subject. Each theme below is a self-contained reference.

## Themes

| Theme | What it covers |
|---|---|
| [[<theme-1>/README\|<theme-1>]] | <one-line scope> |
| ... | ... |

## All docs

![[_index.base]]

## How to use

- Browse by theme (entry: `<theme>/README.md`).
- Use the dashboards (`_index.base` per theme + root) to query by topic, recency, staleness.
- Sources are cited in every page. If a claim has no citation, it's a TODO.

## Conventions

- Themes live under `docs/<theme>/` in `kebab-case`, named after subjects (nouns).
- Each theme has: `README.md`, `_index.base`, plus `overview`, `concepts`, `howto`, `examples`, `gotchas`, `references`.
- Frontmatter on every page: `theme`, `topic`, `last-updated`, `tags`, `sources`.
- Wikilinks resolve within a theme folder; cross-theme uses relative paths (`../<other-theme>/README`).
```

## Root `docs/_index.base` template

```yaml
filters:
  and:
    - file.inFolder(this.file.folder)
    - 'file.ext == "md"'

formulas:
  days_since_update: 'if(last-updated, (today() - date(last-updated)).days, "")'
  is_stale: 'if(last-updated, (today() - date(last-updated)).days > 180, true)'
  is_theme_overview: 'file.basename == "README" && file.folder != this.file.folder'

properties:
  theme:
    displayName: "Theme"
  topic:
    displayName: "Topic"
  last-updated:
    displayName: "Updated"
  formula.days_since_update:
    displayName: "Days Old"

views:
  - type: cards
    name: "Themes"
    filters:
      and:
        - 'formula.is_theme_overview == true'
    order:
      - theme
      - file.name
      - last-updated
    groupBy:
      property: theme
      direction: ASC

  - type: table
    name: "All pages"
    order:
      - file.name
      - theme
      - topic
      - last-updated
      - formula.days_since_update
    groupBy:
      property: file.folder
      direction: ASC

  - type: table
    name: "Stale (>180d)"
    filters:
      and:
        - 'formula.is_stale == true'
    order:
      - file.name
      - theme
      - last-updated
      - formula.days_since_update

  - type: list
    name: "Recently updated"
    limit: 20
    order:
      - file.name
      - theme
      - last-updated
```

## Tools

| Goal | Tool |
|---|---|
| Library / SDK docs | `mcp__plugin_context7_context7__resolve-library-id` + `query-docs` |
| Pull clean markdown from a URL | `obsidian:defuddle` skill (preferred over `WebFetch` for articles) |
| Search the web for sources | `WebSearch` |
| Read existing vault doc | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_read path="docs/<theme>/<page>.md"` |
| Create new vault doc | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_create path="..." content="..."` |
| Append to existing page | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_append path="..." content="..."` |
| Set typed frontmatter | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_property_set name="last-updated" value="2026-05-04" type="date" path="..."` |
| List theme pages | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_files_list folder="docs/<theme>"` |
| Query the dashboard | `mcp__plugin_obsidian-cli-mcp_obsidian-cli__obsidian_base_query path="docs/_index.base" view="..." format=json` |
| Reference base syntax | `obsidian:obsidian-bases` skill |
| Reference markdown syntax | `obsidian:obsidian-markdown` skill |

## Page rules

- **Cite every claim.** Sources go in the page's `sources:` frontmatter and the `## Sources` section. URLs, books, RFCs, papers. No citation → mark as `<TODO: source>`.
- **Synthesize, don't paste.** Defuddle outputs go through your understanding before they hit the page. Pure copy-paste rots when you re-read it.
- **Frontmatter on every page** — `theme`, `topic`, `last-updated` (date), `tags` (list), `sources` (list of URLs/refs). Use `obsidian_property_set` for typed fields.
- **Wikilinks for cross-page nav** — `[[concepts]]`, `[[gotchas]]`. Resolve by filename within the theme folder.
- **Code blocks specify language**. No bare `` ``` ``.
- **Callouts for must-know info** — `> [!warning]`, `> [!info]`, `> [!note]`. Use sparingly.
- **Cross-theme links use relative paths** — `[[../<other-theme>/README]]`.
- **Append over rewrite** for accumulating sections (`gotchas`, `references`, recent-changes). Use `obsidian_append`, not `obsidian_create overwrite=true`.
- **Vault writes via MCP only.** Don't use `Write`/`Edit` for vault paths.

## Delegate to the `documenter` subagent for heavy work

This skill works two ways: parent agent executes inline for small ops, OR delegates to the `documenter` Sonnet subagent for heavier research + writes.

**Inline (parent does directly)**:
- Reading an existing theme's pages
- Appending one gotcha or one source
- Bumping `last-updated`
- Quick lookup against the root `docs/_index.base` to see if a theme exists

**Delegate to `documenter`**:
- **Bootstrap** — creating a new theme from scratch (research + 6+ pages + `_index.base`)
- **Research-heavy extend** — pulling in multiple sources via Context7/defuddle/WebSearch and synthesizing into existing pages
- **Mass refactor** — restructuring a theme's subpages after the topic's mental model evolved
- Anything that'd be >5 MCP write tool calls

How to delegate:

```
Agent({
  description: "Bootstrap docs/<theme>/",
  subagent_type: "documenter",
  prompt: "Bootstrap docs/<theme>/ in the vault. Theme: <subject-noun>. Scope: <one-line>. Subpages: overview, concepts, howto, examples, gotchas, references. Sources to use: <URLs / Context7 lib ids / books>. Use the document-theme skill spec. Cite every claim with sources frontmatter. Synthesize, don't paste."
})
```

The agent runs on Sonnet (cheap + fast for structured-write workloads), invokes this skill itself, executes the bootstrap, returns a structured report listing paths created, sources used, gaps. Don't pass `model: opus` — `documenter` is sonnet by design.

## Anti-patterns

- ❌ Project-specific gotchas in a theme — those go to `semantic-index/<project>/gotchas.md` via `index-project`.
- ❌ One mega-`README.md` instead of split subpages — defeats the linked-doc structure.
- ❌ Wikilinks to non-existent pages — resolve them or remove.
- ❌ Pasting extracted articles verbatim — synthesize.
- ❌ Themes named after sprints, projects, or actions (`q3-research`, `auth-rewrite-2026`) — use subject-nouns.
- ❌ Writing without sources — every page needs the `sources:` frontmatter populated.
- ❌ Stale references that 404 — when re-reading a theme, verify links still resolve.

## Boundaries

- Writes to the **Obsidian vault** at `docs/<theme>/`. Doesn't touch:
  - Repo files (use `Edit`/`Write` for code)
  - `semantic-index/<project>/` (project memory — `index-project` owns it)
  - `Projects/<Customer>/` (manual customer notes)
  - `Reference/<topic>.md` (cross-project gotchas — that's a different namespace, more terse)
- Themes are subjects you'd describe to someone in 5 words ("OAuth", "Postgres replication", "regex"). If your theme name is a sentence, split it.
- This skill is about **durable** knowledge. Ephemeral notes (today's meeting, this sprint's questions) belong in daily notes (`obsidian_daily_append`), not themes.
