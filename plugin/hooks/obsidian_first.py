#!/usr/bin/env python3
"""UserPromptSubmit hook — keep the "Obsidian vault is the primary knowledge
base" directive in Claude's fresh context on every substantive turn.

The plugin's SessionStart hook injects what the vault *contains*, but that
context loses salience as the conversation grows — by message 20 the model has
drifted back to its default instinct of grepping the repo. This hook re-states
the *directive* (not the contents) every turn so it never falls out of the
working set. Trivial confirmation prompts are skipped to avoid pointless noise.

Cost: ~70 tokens per injected turn. Fails silently — never blocks a turn.
"""
import json
import sys

# Pure confirmations / acknowledgements — the directive already landed, skip.
SKIP = {
    "y", "n", "yes", "no", "ok", "okay", "k", "go", "continue", "cont",
    "thanks", "thx", "ty", "stop", "done", "next", "yep", "nope", "sure",
}

DIRECTIVE = (
    "[Standing directive — Obsidian vault = primary knowledge base]\n"
    "Before Grep / Read / Serena on the repo, query the vault first: "
    "obsidian_search or obsidian_search_context (prose) / obsidian_tag "
    "name=\"topic/<x>\" (project microindex). Repo exploration is step 2 — "
    "the vault may already hold the answer.\n"
    "Any new gotcha / decision / convention learned this turn -> capture it "
    "in the vault (semantic-index/<project>/ via index-project, docs/<theme>/, "
    "or Reference/<topic>.md), not just ~/.claude memory files — memory is "
    "project-scoped and not searchable across projects."
)


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return

    prompt = (payload.get("prompt") or "").strip()
    if not prompt or prompt.lower() in SKIP:
        return

    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": DIRECTIVE,
        }
    }))


if __name__ == "__main__":
    main()
