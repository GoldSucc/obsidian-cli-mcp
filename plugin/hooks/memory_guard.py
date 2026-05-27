#!/usr/bin/env python3
"""PostToolUse hook — when Claude writes to the harness project-memory dir
(~/.claude/projects/<slug>/memory/), remind it to mirror the fact into the
Obsidian vault.

Rationale: the harness memory dir is project-scoped and not searchable across
projects, so a memory-only write dead-ends — it never grows the cross-project
knowledge base. The vault does. Double-writing (memory + vault) is fine;
memory-only is the failure mode this hook nudges away from.

Warn-only by design — runs PostToolUse, never blocks or asks. Fails silently.
"""
import json
import re
import sys

# Matches ~/.claude/projects/<slug>/memory/ and ~/.claude/memory/ — the harness
# memory store. Does NOT match the Obsidian vault or .serena/memories/.
MEMORY_RE = re.compile(r"/\.claude/(?:projects/[^/]+/)?memory/")

REMINDER = (
    "[obsidian-cli-mcp] That write landed in the harness memory dir "
    "(~/.claude/projects/<slug>/memory/) — a PROJECT-SCOPED store that is not "
    "searchable from other projects, so on its own it dead-ends and does not "
    "grow the knowledge base.\n"
    "Mirror the same fact into the Obsidian vault now (double-writing is "
    "fine — vault-or-both is correct, memory-only is the thing to avoid):\n"
    "- project-specific gotcha / decision / convention -> "
    "semantic-index/<project>/ (index-project skill)\n"
    "- general / topic knowledge -> docs/<theme>/ (document-theme skill)\n"
    "- cross-cutting (auth, infra, identity, integration) -> "
    "Reference/<topic>.md\n"
    "Use obsidian_create / obsidian_append to do it."
)


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return

    tool_input = payload.get("tool_input") or {}
    path = tool_input.get("file_path") or tool_input.get("path") or ""
    if not path or not MEMORY_RE.search(path):
        return

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": REMINDER,
        }
    }))


if __name__ == "__main__":
    main()
