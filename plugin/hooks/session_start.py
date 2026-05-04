#!/usr/bin/env python3
"""
Inject the current project's semantic index into the Claude session.

Reads `semantic-index/<basename(cwd)>/index.md` from the user's Obsidian vault
via the `obsidian` CLI. If found, dumps the index as additionalContext so Claude
sees the project's known patterns/gotchas at session start. If not found,
nudges Claude toward the `index-project` skill to bootstrap one.

Fail silently on any error so it never blocks a session.
"""
import json
import os
import subprocess
import sys


def main():
    try:
        json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        pass

    cwd = os.getcwd()
    project = os.path.basename(cwd) or ""
    home = os.path.expanduser("~")

    # Don't index the home directory itself or empty paths.
    if not project or cwd in (home, "/"):
        return

    index_path = f"semantic-index/{project}/index.md"

    try:
        result = subprocess.run(
            ["obsidian", "read", f"path={index_path}"],
            capture_output=True,
            text=True,
            timeout=6,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return

    raw = (result.stdout or "").strip()
    # The CLI reports "not found" via stdout with exit 0, prefixed `Error:`.
    missing = result.returncode != 0 or raw.startswith("Error:")
    content = "" if missing else raw

    if content:
        msg = (
            f"## Project Index — {project} (Obsidian vault)\n\n"
            f"Loaded from `semantic-index/{project}/index.md`. "
            f"Use the `index-project` skill to read full topic pages, "
            f"append new findings, or extend the index. "
            f"All persistent knowledge about this project should flow through here.\n\n"
            f"---\n\n{content}"
        )
    else:
        msg = (
            f"## Project Index — {project} (no index yet)\n\n"
            f"No `semantic-index/{project}/index.md` found in the Obsidian vault. "
            f"When you have enough context, invoke the `index-project` skill to "
            f"bootstrap a semantic index for this project. "
            f"Subsequent sessions will auto-load it."
        )

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": msg,
        }
    }))


if __name__ == "__main__":
    main()
