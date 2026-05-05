#!/usr/bin/env python3
"""
Inject the Obsidian vault's knowledge stores into the Claude session.

Three live reads, all derived from vault state (no static index files):

1. **Live theme catalogue** — `obsidian folders folder=docs` enumerates themes.
   Always current, never drifts (filesystem-derived).

2. **Tag taxonomy** — `obsidian tags counts format=tsv` returns the live
   tag/count map. We surface the `theme/*`, `topic/*`, and `project/*`
   namespaces so Claude can drill in by tag (`obsidian_tag name=topic/<x>
   verbose`) instead of re-reading every README.

3. **Project semantic index** — `semantic-index/<basename(cwd)>/index.md`. May
   not match the actual project name (cwd could be a subdir, dir could be
   generic like `code/`). When the basename guess fails, the injected guidance
   tells Claude to detect the real project name from CLAUDE.md / README.md /
   package.json / pyproject.toml / go.mod / Cargo.toml / git remote / user
   mentions, then look up the index via `obsidian_read`.

Fail silently on any error — the hook must never block a session.
"""
import json
import os
import subprocess
import sys
from collections import defaultdict


def cli(*args: str, timeout: int = 6) -> str:
    """Run obsidian CLI, return stripped stdout. Empty string on any failure."""
    try:
        result = subprocess.run(
            ["obsidian", *args],
            capture_output=True,
            text=True,
            timeout=timeout,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return ""
    raw = (result.stdout or "").strip()
    if result.returncode != 0 or raw.startswith("Error:"):
        return ""
    return raw


def list_themes() -> list[str]:
    """Direct children of docs/ — one folder per theme."""
    raw = cli("folders", "folder=docs")
    if not raw:
        return []
    themes: list[str] = []
    for line in raw.splitlines():
        line = line.strip()
        if not line or line == "docs":
            continue
        if not line.startswith("docs/"):
            continue
        rest = line[len("docs/"):]
        if "/" in rest:
            continue
        themes.append(rest)
    themes.sort()
    return themes


def collect_tags() -> dict[str, list[tuple[str, int]]]:
    """Group tags by top-level namespace. Returns {namespace: [(tag, count), ...]}."""
    raw = cli("tags", "counts", "sort=count", "format=tsv")
    if not raw:
        return {}
    grouped: dict[str, list[tuple[str, int]]] = defaultdict(list)
    for line in raw.splitlines():
        if "\t" not in line:
            continue
        tag, count = line.split("\t", 1)
        tag = tag.lstrip("#").strip()
        try:
            n = int(count.strip())
        except ValueError:
            continue
        if "/" in tag:
            ns, _ = tag.split("/", 1)
        else:
            ns = "_flat"
        grouped[ns].append((tag, n))
    return grouped


def fmt_tag_list(tags: list[tuple[str, int]], limit: int = 20) -> str:
    if not tags:
        return ""
    head = tags[:limit]
    rest = len(tags) - limit
    parts = [f"`{t}` ({c})" for t, c in head]
    line = ", ".join(parts)
    if rest > 0:
        line += f", … (+{rest} more)"
    return line


def truncate(s: str, limit: int = 4000) -> str:
    if len(s) <= limit:
        return s
    return s[:limit] + "\n\n…(truncated — read with `obsidian_read` for the rest)"


def main():
    try:
        json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        pass

    cwd = os.getcwd()
    project_guess = os.path.basename(cwd) or ""
    home = os.path.expanduser("~")

    if not project_guess or cwd in (home, "/"):
        return

    sections: list[str] = []

    # 1. Live theme catalogue
    themes = list_themes()
    if themes:
        theme_list = "\n".join(f"- `[[docs/{t}/README]]` — `{t}`" for t in themes)
        sections.append(
            "## General knowledge themes (vault `docs/`)\n\n"
            f"{len(themes)} theme(s) currently documented (live folder listing — always current).\n"
            "To see what a theme covers: `obsidian_read path=\"docs/<theme>/README.md\"`.\n"
            "To extend or bootstrap: `document-theme` skill or `documenter` agent.\n\n"
            f"{theme_list}"
        )
    else:
        sections.append(
            "## General knowledge themes (vault `docs/`)\n\n"
            "No themes documented yet. When the user asks to research or document a topic, invoke `document-theme`."
        )

    # 2. Tag taxonomy — drives content/topic discovery
    grouped = collect_tags()
    if grouped:
        tax_lines: list[str] = []
        # Show the meaningful namespaces first.
        for ns, label in (
            ("theme", "Themes (`#theme/*`)"),
            ("topic", "Topics (`#topic/*`)"),
            ("project", "Projects (`#project/*`)"),
            ("subsystem", "Subsystems (`#subsystem/*`)"),
            ("kind", "Kinds (`#kind/*`)"),
            ("docs", "Doc anchors (`#docs/*`)"),
        ):
            tags = grouped.get(ns, [])
            if not tags:
                continue
            tax_lines.append(f"### {label}\n{fmt_tag_list(tags, limit=20)}")
        # Refs are mechanical (one per indexed object) — surface count + top examples.
        ref_tags = grouped.get("ref", [])
        if ref_tags:
            top_refs = ", ".join(f"`{t}` ({c})" for t, c in ref_tags[:5])
            tax_lines.append(
                f"### References (`#ref/*`)\n"
                f"{len(ref_tags)} ref tag(s) across the vault — `#ref/<NAME>` is the dependency-graph namespace. "
                f"Each microindex note carries `#ref/<self>` plus one `#ref/<dep>` per outgoing reference. "
                f"Query a single object's call graph with `obsidian_tag name=\"ref/<NAME>\" verbose`. "
                f"Top: {top_refs}."
            )
        if tax_lines:
            sections.append(
                "## Tag taxonomy (live, vault-wide)\n\n"
                "Drill into any tag below for filtered content. Discovery patterns:\n"
                "- Files for a topic tag: `obsidian_tag name=\"topic/<x>\" verbose`\n"
                "- **Dependency / call graph**: `obsidian_tag name=\"ref/<OBJECT_NAME>\" verbose` — returns the object's microindex note (self-tag) AND every unit that references it (outgoing). Use this BEFORE grepping the codebase.\n"
                "- Content search scoped to docs: `obsidian_search query=\"<text>\" path=\"docs\"`\n"
                "- With surrounding context: `obsidian_search_context query=\"<text>\" path=\"docs\"`\n"
                "- Outline a page: `obsidian_outline path=\"docs/<theme>/<page>.md\" format=md`\n\n"
                + "\n\n".join(tax_lines)
            )

    # 3. Project semantic index
    project_index = cli("read", f"path=semantic-index/{project_guess}/index.md")
    if project_index:
        sections.append(
            f"## Project semantic index — `{project_guess}`\n\n"
            f"Loaded from `semantic-index/{project_guess}/index.md`. "
            f"All persistent project knowledge flows through here. "
            f"Use `index-project` skill (or `documenter` agent) to extend.\n\n"
            f"---\n\n{truncate(project_index)}"
        )
    else:
        sections.append(
            f"## Project semantic index (no match for `{project_guess}`)\n\n"
            f"No `semantic-index/{project_guess}/index.md` found — the cwd basename `{project_guess}` may not be the actual project name.\n\n"
            "**Detect the real project name** from any of:\n"
            "- `CLAUDE.md` (project section / first heading)\n"
            "- `README.md` (title)\n"
            "- `package.json` `name`, `pyproject.toml` `[project].name`, `go.mod` `module`, `Cargo.toml` `[package].name`\n"
            "- Git remote (`git remote -v` → repo name)\n"
            "- User mentions (\"working on X\", \"the X project\")\n\n"
            "Alternative: search the `#project/*` taxonomy above for an existing match.\n\n"
            "Once the name is clear, look up `semantic-index/<name>/index.md` via `obsidian_read`. "
            "If still missing, invoke `index-project` (or dispatch `documenter`) to bootstrap."
        )

    # 4. Microindex (semantic LSP) — advertise query workflow and current state
    micro_files = cli("files", f"folder=semantic-index/{project_guess}/index", "ext=md")
    micro_count = 0
    micro_kinds: dict[str, int] = defaultdict(int)
    if micro_files:
        for line in micro_files.splitlines():
            line = line.strip()
            if not line.endswith(".md"):
                continue
            micro_count += 1
            # path: semantic-index/<project>/index/<kind>/<name>.md  →  kind = third segment from base
            parts = line.split("/")
            if len(parts) >= 4 and parts[0] == "semantic-index" and parts[2] == "index":
                kind = parts[3] if len(parts) >= 5 else "_root"
                micro_kinds[kind] += 1

    if micro_count > 0:
        kind_summary = ", ".join(f"{k} ({n})" for k, n in sorted(micro_kinds.items(), key=lambda x: -x[1]))
        sections.append(
            f"## Microindex (semantic LSP) — `{project_guess}` has {micro_count} indexed unit(s)\n\n"
            f"Live count from `semantic-index/{project_guess}/index/`. Breakdown: {kind_summary}.\n\n"
            "**Query the microindex BEFORE exploring blindly.** Each indexed unit — source file, ABAP "
            "object, config key, env var, endpoint, DB table, IaC module, etc. — is a tiny note with "
            "`#topic/*` tags + anchors, designed for retrieval, not reading.\n\n"
            "Search workflow:\n\n"
            f"1. **Tag-first** — `obsidian_tag name=\"topic/<x>\" verbose path=\"semantic-index/{project_guess}/index\"` returns every indexed unit touching the topic.\n"
            f"2. **Read the matching microindex notes** — `obsidian_read path=\"semantic-index/{project_guess}/index/<kind>/<name>.md\"` shows the anchors and the fetchable `ref`.\n"
            "3. **Fetch the actual unit** — use the `ref` from the note. For source files: `Read path=...:lines`. For ABAP: `mcp__plugin_vsp_sap-adt__GetSource`. For configs: `Read` the file, find the key. The reference taxonomy in the note tells you which tool fits.\n\n"
            "Other entry points:\n\n"
            f"- Content search: `obsidian_search_context query=\"<text>\" path=\"semantic-index/{project_guess}/index\"` — returns matching anchor lines with surrounding context.\n"
            f"- Pivoted dashboard: open `semantic-index/{project_guess}/index/_index.base` in Obsidian to filter by `kind`, `module`, or `theme`.\n"
            f"- List by kind: `obsidian_files_list folder=\"semantic-index/{project_guess}/index/<kind>\"` (e.g. `kind=class`, `kind=cds`, `kind=config`, `kind=endpoint`).\n\n"
            "**Use this BEFORE `Grep`/`Read` on the repo OR before MCP-querying ABAP objects directly.** The microindex was built precisely so Claude doesn't have to re-discover the project every session. If a query returns nothing, fall back AND propose extending the microindex."
        )
    else:
        sections.append(
            f"## Microindex (semantic LSP) — none yet for `{project_guess}`\n\n"
            f"No `semantic-index/{project_guess}/index/` folder. The microindex is the per-unit retrieval layer — "
            "one tiny note per indexable unit (source file, ABAP class/CDS/BDef via MCP, config key, env var, "
            "HTTP endpoint, DB table, IaC module — anything fetchable) with `#topic/*` tags and anchors. "
            "Without it, Claude has to re-discover the project every session.\n\n"
            "**Build it once, query it forever:**\n\n"
            "```\n"
            "Agent({\n"
            f"  description: \"Microindex {project_guess}\",\n"
            "  subagent_type: \"microindexer\",\n"
            f"  prompt: \"Build microindex for project {project_guess}. Sources: <filesystem dirs and/or ABAP packages and/or config files and/or DB schemas — whatever applies>. Reuse #topic/* tags from semantic-index/{project_guess}/. Cross-link to topic pages and docs/<theme>/ where applicable.\"\n"
            "})\n"
            "```\n\n"
            "Microindexer runs on Haiku — cheap mass scanning across whatever unit sources the project has. "
            "Required prerequisite: topic pages must exist first (run `index-project` skill or dispatch `documenter` agent for step 1)."
        )

    msg = "\n\n".join(sections)

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": msg,
        }
    }))


if __name__ == "__main__":
    main()
