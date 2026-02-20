#!/bin/sh
# tztoolbox installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tzoght/tztoolbox/main/install.sh | sh
#   — or —
#   git clone https://github.com/tzoght/tztoolbox.git && cd tztoolbox && sh install.sh
set -e

REPO_URL="https://github.com/tzoght/tztoolbox.git"
CURSOR_HOME="${HOME}/.cursor"
REPO_CURSOR=".cursor"

cleanup() {
  if [ -n "${TMPDIR_CREATED:-}" ] && [ -d "$TMPDIR_CREATED" ]; then
    rm -rf "$TMPDIR_CREATED"
  fi
}
trap cleanup EXIT INT TERM

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: '$1' is required but not found." >&2
    exit 1
  fi
}

# Determine source directory: local clone or remote fetch
if [ -d "$REPO_CURSOR/commands" ] && [ -d "$REPO_CURSOR/rules" ]; then
  SRC="."
  echo "Detected local clone — installing from working tree."
else
  need_cmd git
  TMPDIR_CREATED="$(mktemp -d)"
  echo "Cloning tztoolbox into temp directory..."
  git clone --depth 1 "$REPO_URL" "$TMPDIR_CREATED/tztoolbox" 2>&1 | sed 's/^/  /'
  SRC="$TMPDIR_CREATED/tztoolbox"
fi

mkdir -p "$CURSOR_HOME/commands" "$CURSOR_HOME/rules" "$CURSOR_HOME/skills"

installed_commands=0
installed_rules=0
installed_skills=0

for f in "$SRC/$REPO_CURSOR"/commands/*.md; do
  [ -f "$f" ] || continue
  cp "$f" "$CURSOR_HOME/commands/$(basename "$f")"
  installed_commands=$((installed_commands + 1))
done

for f in "$SRC/$REPO_CURSOR"/rules/*; do
  [ -f "$f" ] || continue
  cp "$f" "$CURSOR_HOME/rules/$(basename "$f")"
  installed_rules=$((installed_rules + 1))
done

for d in "$SRC/$REPO_CURSOR"/skills/*/; do
  [ -d "$d" ] || continue
  cp -r "${d%/}" "$CURSOR_HOME/skills/"
  installed_skills=$((installed_skills + 1))
done

echo ""
echo "tztoolbox installed successfully!"
echo "  Commands : $installed_commands -> $CURSOR_HOME/commands/"
echo "  Rules    : $installed_rules -> $CURSOR_HOME/rules/"
echo "  Skills   : $installed_skills -> $CURSOR_HOME/skills/"
echo ""
echo "Goodies are now available in all Cursor projects."
echo "If they don't appear in Cursor Settings, restart Cursor once."
