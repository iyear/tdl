---
name: tdl
description: Operate the tdl Telegram CLI on the user's behalf with TDL_AGENT=1, QR-only login, per-conversation namespace and proxy confirmation, and confirmation before every write operation. Use for tdl login, downloads, uploads, forwarding, chat exports, backups, migration, extensions, and troubleshooting; do not use for Telegram work unrelated to tdl.
---

# tdl

Complete Telegram file and message tasks by operating `tdl` in the user's environment. Run read-only commands after the session preflight. Run state-changing commands only after the operation-specific choices are complete and the user explicitly confirms that exact write operation.

## Agent marker

Every `tdl` process launched by the agent must inherit the environment variable `TDL_AGENT=1`. This is an unconditional invariant for all invocations, including `version`, `help`, read-only queries, dry-runs, interactive commands, extensions, retries, and commands inside pipelines or wrappers.

Set the variable on the child process through the execution tool's environment facility when available. Otherwise, scope it to the individual shell command:

```bash
TDL_AGENT=1 tdl <global-flags> <command>
```

Override any existing `TDL_AGENT` value with `1`. Do not persist the variable in the user's shell profile or global environment. When only preparing a command for the user to run personally, omit the marker because that process is not agent-launched.

## Login policy

The agent may log in only with QR authentication. Always pass `--type qr` (or `-T qr`) explicitly. Never invoke the default desktop-session import, `--type code`, the deprecated `--code` flag, or any other login mode, even if the user requests it. Explain that this skill supports QR login only.

Before constructing a login command, ask whether the Telegram account requires a passcode or 2FA password:

- If required, obtain it before requesting write confirmation and include it in the initial command with `--passcode`. Never start login first and enter the passcode at a later interactive prompt.
- If the user explicitly says no passcode is required, omit `--passcode`.
- If the requirement is unknown, do not start login until the user resolves it.

Treat the passcode as sensitive. Pass it exactly once to the command, do not repeat it in commentary or summaries, and redact its value when previewing the command for write confirmation.

With `TDL_AGENT=1`, QR login prints a one-time browser URL for the QR image. Present that URL only to the user and keep the process running while they scan it.

## Session preflight

Before the first `tdl` command in every conversation, determine whether the user has explicitly confirmed both of these values in the current conversation:

- `ns`: the exact tdl account namespace.
- `proxy`: either an exact proxy URL or an explicit choice to connect directly without a proxy.

Do not infer either value from tdl defaults, environment variables, local files, a previous conversation, or an account that appears likely. If either value is missing, ask the user before running any `tdl` command. Once confirmed, reuse it throughout the current conversation without asking again unless the user changes it.

Pass the confirmed namespace and proxy consistently on every tdl invocation in the workflow. For a confirmed direct connection, omitting `--proxy` or passing an explicit empty value is acceptable. Other [global config](references/guide/global-config.md) values may use their documented defaults without confirmation unless the user supplies different values.

## Operate the CLI

1. Establish the requested outcome and complete the session preflight above.
2. Read the smallest relevant set of bundled references below. Derive flags and behavior from those references rather than memory.
3. Check that `tdl` is available and run `tdl version`. Consult `tdl <command> --help` when the installed version may differ from the bundled documentation. Apply the agent marker to these checks and every later invocation.
4. Resolve operation-specific input. Preserve the user's paths, links, chat identifiers, destinations, ranges, and requested options. For login, enforce the QR-only and passcode rules above instead of asking the user to choose a login mode. Ask when another user-visible choice cannot be safely inferred; do not ask about other global-config defaults merely because they exist.
5. Build the simplest command that achieves the requested outcome. Quote paths, proxy URLs, and expressions for the active shell. For a read-only command, run it directly. For a write command, summarize its exact scope and effects, request confirmation, and run it only after the user confirms.
6. For an interactive prompt, show the prompt and its choices to the user and wait for an explicit selection. Never press Enter to accept a default or select an option on the user's behalf. After the user chooses, relay only that exact response to the running process. QR scanning remains a user-only action. Never answer an interactive passcode or 2FA prompt; stop and restart only after constructing a QR command with `--passcode` and obtaining a new write confirmation.
7. Monitor long-running commands instead of treating process startup as success. Keep interactive sessions open while waiting for the user, and stop retrying when the same failure needs new input or authority.
8. Verify the exit status and relevant artifact or Telegram-side result. Report output paths, completed work, skipped items, and failures.

## Write operations

Treat a command as a write operation when its intended behavior changes local files, tdl session or storage data, installed software, or Telegram-side state. This includes login, downloads, exports, backups, uploads, forwards, message edits, migration, recovery, and extension installation, upgrade, or removal.

Before every write operation:

1. Show the target and scope, including local paths, Telegram source and destination, message or file range, and any overwrite, deletion, logout, or third-party-code effect that applies.
2. Ask the user to confirm execution. The original task request, session preflight, operation-specific choices, silence, or approval of a dry-run do not count as this confirmation.
3. After an explicit affirmative response, execute the confirmed command on the user's behalf and monitor it through completion or a user-required prompt.

A confirmation covers only the exact command or clearly enumerated batch presented to the user. Obtain a new confirmation before changing scope, adding a recipient or file, enabling a destructive flag, or retrying in a way that may duplicate or overwrite work. Upload and forward always require this write confirmation even when their source and destination were already selected.

## Reference routing

- Installation and first login: [installation](references/getting-started/installation.md) and [quick start and login](references/getting-started/quick-start.md). Ignore the non-QR login methods in the quick-start reference. Read [shell completion](references/getting-started/shell-completion.md) only when the user requests completion setup.
- Account namespace, proxy, storage, NTP, reconnect, logging, pool, and delay flags: [global config](references/guide/global-config.md). These flags apply only to the current invocation; they are not persisted.
- Environment-variable equivalents: [environment variables](references/more/env.md). Flags take precedence, and `TDL_STORAGE` uses a JSON-object format rather than the CLI flag format.
- Download from message links or exported JSON: [download](references/guide/download.md). For JSON generation, also read [export messages](references/guide/tools/export-messages.md); for custom filenames, read [template guide](references/guide/template.md).
- Upload files or directories, choose a destination, route by expression, set captions, or filter extensions: [upload](references/guide/upload.md).
- Forward, clone, route, edit, dry-run, reorder, or silently send messages: [forward](references/guide/forward.md).
- Discover chats and topic IDs: [list chats](references/guide/tools/list-chats.md). Accepted chat identifiers are summarized in [chat examples](references/snippets/chat.md).
- Export messages or members: [export messages](references/guide/tools/export-messages.md) and [export members](references/guide/tools/export-members.md).
- Back up, recover, or move namespace storage: [migration](references/guide/migration.md).
- Install, run, upgrade, list, or remove experimental extensions: [extensions](references/guide/extensions.md).
- Expressions used by filters, routers, captions, and edits: [expression guide](references/reference/expr.md). Use the command's `-` expression value, where documented, to list the fields available to that command.
- Data locations and diagnostics: [data locations](references/more/data.md) and [troubleshooting](references/more/troubleshooting.md).
- Supported Telegram message-link forms: [message-link examples](references/snippets/link.md).

## Operational constraints

- A namespace represents one Telegram account. Use only the namespace confirmed in the current conversation and keep the same `-n/--ns` value across the workflow.
- Global flags apply only to one invocation. Repeat the confirmed namespace and proxy on every command. Storage, NTP, reconnect timeout, debug logging, pool size, delay, and progress settings may retain their documented defaults unless the user requests otherwise.
- Require an explicit destination for upload and forward; never silently rely on Saved Messages. Login mode is not a choice: always use QR. Require the user to decide passcode requirements, account scanning, overwrite behavior, destructive flags, and other consequential operation-specific options when they apply.
- Confirmation of a destination or another operation-specific choice is not permission to execute the write. Present the completed upload, forward, login, download, export, backup, migration, recovery, or extension command for a separate execution confirmation.
- Thread count, task concurrency, ordering, resume behavior, and other non-global command options may use documented defaults when the user's request does not depend on them. Avoid excessive parallelism or bulk transfers that can trigger Telegram flood control.
- Scope uploads, forwards, exports, and downloads exactly to the chats, links, ranges, and files selected by the user. A successful dry-run does not authorize the real operation.
- `tdl recover` overwrites existing namespace data, and upload `--rm` deletes local files after successful upload. Execute them only after the user explicitly confirms the stated effect.
- Treat login material, backup files, `~/.tdl`, proxy credentials, and `GITHUB_TOKEN` as sensitive. Never echo or expose their contents.
- Do not combine mutually exclusive options documented for a command, including upload routing `--to` with `--chat` or `--topic`, or file-extension whitelists with blacklists.
- Treat extension installation and upgrades as execution of third-party software. Run them only for the exact extension source requested by the user.
- Place global flags before an extension command and extension-specific flags after it: `tdl <global-flags> <extension-name> <extension-flags>`.
