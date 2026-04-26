# Reference: MCP + 1Password CLI

## Secret reference syntax

Format:

```text
op://<vault>/<item>/<field>
```

Use **fictional** names in docs (e.g. `op://Engineering/MCP Service Account/credential`). Real paths belong only in the user’s local config or vault.

## Option A: `op run` + `--env-file` wrapping the server command

Cursor MCP config often uses a single `command` plus `args`. To avoid plaintext tokens, point `op run` at an **env file** whose values are secret references, then start the server after `--`.

**Local env file** (e.g. `.cursor/mcp-secrets.env` — keep **gitignored**; commit a `.example` template with fake `op://` paths only if helpful):

```text
# Example only — use your real vault/item/field labels
OP_SERVICE_ACCOUNT_TOKEN=op://Engineering/MCP Service Account/credential
```

**MCP entry** (paths may need to be absolute depending on how Cursor spawns the server):

```json
{
  "mcpServers": {
    "example-vault-tools": {
      "command": "op",
      "args": [
        "run",
        "--env-file",
        ".cursor/mcp-secrets.env",
        "--",
        "npx",
        "-y",
        "@example/community-1password-mcp"
      ]
    }
  }
}
```

**Notes:**

- Use [`--env-file`](https://developer.1password.com/docs/cli/reference/commands/run) per official CLI docs; do not put real tokens in JSON.
- `--no-masking` (optional on `op run`) helps local debugging; omit if policy requires masking.
- Replace `@example/community-1password-mcp` with whatever server the user chose after review; packages differ in env var names.
- **Community servers are unofficial** — verify package name and behavior before use.

## Option B: env file resolved by `op run`

If the MCP server reads many variables, a local env file whose **values** are `op://...` references can be used with `op run --env-file=path -- command` (see [1Password CLI run](https://developer.1password.com/docs/cli/reference/commands/run)). Keep that file **out of git** if it contains real vault paths, or use a committed template with placeholders and document copying to a gitignored path.

## Illustrative `mcp.json` (placeholders only)

Assumes `.cursor/mcp-secrets.env` contains lines like `API_KEY=op://VaultName/ItemName/field_label`.

```json
{
  "mcpServers": {
    "myServer": {
      "command": "op",
      "args": [
        "run",
        "--env-file",
        ".cursor/mcp-secrets.env",
        "--",
        "node",
        "/absolute/path/to/mcp-server/dist/index.js"
      ]
    }
  }
}
```

## Common env var names (community servers)

Servers differ. Define **their** documented variables in the `--env-file` (values = `op://…` references):

| Typical purpose | Example env names (varies by server) |
|-----------------|--------------------------------------|
| 1Password service account token | `OP_SERVICE_ACCOUNT_TOKEN`, `ONEPASSWORD_SERVICE_ACCOUNT_TOKEN` |
| 1Password Connect token / URL | `OP_CONNECT_TOKEN`, `OP_CONNECT_HOST` |

Always read the **specific server’s README**; do not assume variable names.

## Troubleshooting

| Symptom | Things to check |
|---------|------------------|
| `op` fails immediately | Signed in? (`op signin` or service account token valid?) Correct account? |
| MCP server starts but auth fails | Service account has vault/item access? Item field label matches reference? |
| Works locally, fails in CI | CI needs injected token (e.g. GitHub Actions secret → `op run` or short-lived token pattern per org policy). |
| Secrets in logs | Avoid `--no-masking` in shared logs; reduce server debug level. |

## Links

- [1Password CLI](https://developer.1password.com/docs/cli)
- [Service accounts](https://developer.1password.com/docs/service-accounts/)
- [Securing MCP servers with 1Password](https://1password.com/blog/securing-mcp-servers-with-1password-stop-credential-exposure-in-your-agent)
