---
name: 1password-mcp
description: >-
  Securely connect Cursor MCP to 1Password using the 1Password CLI (op run,
  secret references) and optional third-party MCP servers. Use when the user
  asks about MCP + 1Password, Cursor MCP config, service accounts for agents,
  injecting tokens without plaintext in mcp.json, or bridging vault access to AI tools.
---

# MCP bridge to 1Password (Cursor)

This skill describes **patterns**, not a bundled MCP server. Prefer **1Password CLI** injection so secrets never live in git or plain `mcp.json`. Official background: [Securing MCP servers with 1Password](https://1password.com/blog/securing-mcp-servers-with-1password-stop-credential-exposure-in-your-agent) and [Where MCP fits and where it doesn’t](https://1password.com/blog/where-mcp-fits-and-where-it-doesnt).

## When to use what

| Need | Approach |
|------|----------|
| MCP server needs an API token or password in **env** | Run the **MCP server process** (or Cursor) under `op run` so env values resolve from vault fields; see [references/reference.md](references/reference.md). |
| Docs or config examples for the user | Use **placeholders** only: `op://VaultName/ItemName/field` — never real vault/item names from the user’s account unless they pasted them first. |
| Agent should **read or manage** vault items via MCP | **Optional**: community MCP servers (not first-party 1Password products). User must **opt in**, **review the server code**, and use a **least-privilege 1Password service account**. Token still supplied only via `op run` / secret reference, not committed files. |

## Cursor wiring

- **Project MCP:** many setups use [`.cursor/mcp.json`](references/reference.md) at the repo root (confirm in [Cursor documentation](https://cursor.com/docs) for your version).
- **User-wide MCP:** Cursor may also support global MCP config in user settings; prefer project-level for team-shared **patterns** (still without plaintext secrets).

**Shape:** MCP entries use a `command` and `args` (and sometimes `env`). The safe pattern is either:

1. **`command` = `op`**, **`args` = `run`, `--env-file`, `<path>`, `--`, `<server-binary>`, ...** so 1Password resolves `op://` references from the file into the subprocess env (see [1Password CLI: secrets in environment variables](https://developer.1password.com/docs/cli/secrets-environment-variables)), or  
2. A **small wrapper script** that exports secret-reference vars and runs `op run -- …` (mind [subshell expansion order](https://developer.1password.com/docs/cli/secrets-environment-variables)) so JSON stays free of tokens.

Do not put service account tokens or passwords as literal strings in JSON or markdown you commit.

## Operational checklist

1. **Automation:** use a [1Password service account](https://developer.1password.com/docs/service-accounts/) with access to **only** the vaults needed for MCP; avoid broad admin tokens.
2. **Interactive dev:** signed-in `op` session can work locally; CI should use service account tokens via injected env, never hardcoded in workflow YAML.
3. **Rotation:** if a token appears in chat, a PR, or a leak, **rotate** it in 1Password and update references; assume compromise.
4. **`.env`:** do not commit real `.env` files; if using `op run` with a template, keep the template in git with `op://...` references only where supported, or document a local-only `.env` in `.gitignore`.

## Agent rules

- **Never** paste vault passwords, OTPs, or private keys into the chat, commits, PR descriptions, or issues.
- When editing **MCP or app config** for the user, use **generic placeholder** vault/item/field names unless the user explicitly provided real paths.
- Prefer pointing the user to **1Password Developer** docs for SDKs and service accounts rather than inventing vault semantics.
- Treat **third-party MCP servers** as **untrusted code** until the user reviews them; do not recommend enabling them by default.

## Snippets and examples

Copy-paste JSON patterns, env var names, and troubleshooting: [references/reference.md](references/reference.md).
