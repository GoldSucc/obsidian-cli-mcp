# Tool Reference

All tools are stdio MCP tools, prefix `mcp__obsidian-cli__`. Names below omit the prefix.

## Files & content

| Tool | CLI | Purpose |
|---|---|---|
| `obsidian_read` | `read` | Full contents of a note |
| `obsidian_create` | `create` | New note (name or path) |
| `obsidian_append` | `append` | Add content to end |
| `obsidian_prepend` | `prepend` | Add content to start |
| `obsidian_move` | `move` | Move/rename via path |
| `obsidian_rename` | `rename` | Rename in place |
| `obsidian_delete` | `delete` | Trash by default; `permanent=true` skips trash |
| `obsidian_file_info` | `file` | Metadata (size, dates) |
| `obsidian_files_list` | `files` | List files; folder/ext filter |
| `obsidian_wordcount` | `wordcount` | Words and/or characters |

## Folders

| Tool | CLI |
|---|---|
| `obsidian_folder` | `folder` |
| `obsidian_folders` | `folders` |

## Daily

| Tool | CLI |
|---|---|
| `obsidian_daily` | `daily` |
| `obsidian_daily_read` | `daily:read` |
| `obsidian_daily_append` | `daily:append` |
| `obsidian_daily_prepend` | `daily:prepend` |
| `obsidian_daily_path` | `daily:path` |

## Search

| Tool | CLI | Notes |
|---|---|---|
| `obsidian_search` | `search` | Returns matching files |
| `obsidian_search_context` | `search:context` | Returns matching lines + surrounding context |
| `obsidian_search_open` | `search:open` | Opens search panel in app |

## Tasks

| Tool | CLI | Purpose |
|---|---|---|
| `obsidian_task` | `task` | Show or update one task (toggle / done / todo / status char) |
| `obsidian_tasks` | `tasks` | List & filter |

## Properties (frontmatter)

| Tool | CLI |
|---|---|
| `obsidian_properties` | `properties` |
| `obsidian_property_read` | `property:read` |
| `obsidian_property_set` | `property:set` |
| `obsidian_property_remove` | `property:remove` |

## Tags & links

| Tool | CLI |
|---|---|
| `obsidian_tag` | `tag` |
| `obsidian_tags` | `tags` |
| `obsidian_aliases` | `aliases` |
| `obsidian_links` | `links` |
| `obsidian_backlinks` | `backlinks` |
| `obsidian_unresolved` | `unresolved` |
| `obsidian_orphans` | `orphans` |
| `obsidian_deadends` | `deadends` |
| `obsidian_outline` | `outline` |

## Random / templates / bases / bookmarks

| Tool | CLI |
|---|---|
| `obsidian_random` | `random` |
| `obsidian_random_read` | `random:read` |
| `obsidian_templates` | `templates` |
| `obsidian_template_read` | `template:read` |
| `obsidian_template_insert` | `template:insert` |
| `obsidian_bases` | `bases` |
| `obsidian_base_views` | `base:views` |
| `obsidian_base_query` | `base:query` |
| `obsidian_base_create` | `base:create` |
| `obsidian_bookmark` | `bookmark` |
| `obsidian_bookmarks` | `bookmarks` |

## History

| Tool | CLI |
|---|---|
| `obsidian_history` | `history` |
| `obsidian_history_list` | `history:list` |
| `obsidian_history_read` | `history:read` |
| `obsidian_history_restore` | `history:restore` |
| `obsidian_history_open` | `history:open` |
| `obsidian_diff` | `diff` |

## Sync

| Tool | CLI |
|---|---|
| `obsidian_sync` | `sync` |
| `obsidian_sync_status` | `sync:status` |
| `obsidian_sync_deleted` | `sync:deleted` |
| `obsidian_sync_history` | `sync:history` |
| `obsidian_sync_read` | `sync:read` |
| `obsidian_sync_restore` | `sync:restore` |
| `obsidian_sync_open` | `sync:open` |

## Vault & app

| Tool | CLI |
|---|---|
| `obsidian_vault_info` | `vault` |
| `obsidian_vaults` | `vaults` |
| `obsidian_reload` | `reload` |
| `obsidian_restart` | `restart` |
| `obsidian_version` | `version` |

## Workspace

| Tool | CLI |
|---|---|
| `obsidian_workspace` | `workspace` |
| `obsidian_tabs` | `tabs` |
| `obsidian_tab_open` | `tab:open` |
| `obsidian_recents` | `recents` |

## Commands & hotkeys

| Tool | CLI | Notes |
|---|---|---|
| `obsidian_command` | `command` | Execute any Obsidian command by id (incl. plugin-contributed) |
| `obsidian_commands` | `commands` | List available command ids |
| `obsidian_hotkey` | `hotkey` | Lookup one |
| `obsidian_hotkeys` | `hotkeys` | List all |

## Escape hatch

| Tool | Purpose |
|---|---|
| `obsidian_run` | Run any CLI command not typed above. Pass `command`, `params{}`, `flags[]`, `vault?`. |

> [!info] Excluded from typed surface
> Plugin/theme/snippet management (`plugin:*`, `theme:*`, `snippet:*`), dev tooling (`dev:*`, `eval`, `devtools`) — reachable via `obsidian_run`. Kept out to avoid bloating the LLM-visible tool catalogue with niche commands.

## File targeting (most tools accept these)

- `vault=<name>` — defaults to `OBSIDIAN_DEFAULT_VAULT` env or most-recently focused
- `file=<name>` — wikilink-style; resolves by filename, ignoring path/extension
- `path=<folder/note.md>` — exact path from vault root
- Omit both `file` and `path` for current active file (where supported)
