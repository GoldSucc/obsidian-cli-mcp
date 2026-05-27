#!/usr/bin/env python3
"""PreToolUse hook — fire ONCE per session, the first time Claude reaches for
repo exploration (Grep / Glob / Serena search), to catch the "explore the repo
first" instinct at the exact moment it happens.

A per-session sentinel file (keyed by session_id, in the system temp dir)
guarantees the reminder is injected at most once per session — precise
interception without becoming noise on every subsequent search.

Never blocks the tool — only injects context. Fails silently.
"""
import json
import os
import sys
import tempfile

REMINDER = (
    "[obsidian-cli-mcp] First repo exploration this session. The Obsidian "
    "vault is the primary knowledge base — if you have not queried it yet "
    "this session, do that FIRST: obsidian_search / obsidian_search_context "
    "for prose, obsidian_tag name=\"topic/<x>\" for the project microindex. "
    "It may already answer this without touching the repo. The tool will "
    "still run — this is a one-time-per-session reminder."
)


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return

    session_id = payload.get("session_id") or "nosession"
    safe = "".join(c for c in session_id if c.isalnum() or c in "-_") or "x"
    sentinel = os.path.join(
        tempfile.gettempdir(), f"claude-obsidian-explore-{safe}"
    )

    if os.path.exists(sentinel):
        return
    try:
        with open(sentinel, "w") as fh:
            fh.write("1")
    except OSError:
        # Could not write the sentinel — inject once anyway, do not loop on it.
        pass

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "additionalContext": REMINDER,
        }
    }))


if __name__ == "__main__":
    main()
