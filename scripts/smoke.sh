#!/usr/bin/env bash
#
# Copyright 2026 Cloudmanic Labs, LLC. All rights reserved.
# Date: 2026-06-22
#
# smoke.sh — end-to-end smoke test of the harbor CLI against a live (dev) API.
#
# It exercises the headline flow: login → create notebook → create note
# (Markdown) → tag → search → export → logout, asserting each step, using an
# isolated HOME so it never touches your real ~/.config/harbor.
#
# This needs a reachable API and a valid account, so it is NOT part of the
# no-network unit-test suite — run it manually against a local/dev server:
#
#   HARBOR_API_URL=http://localhost:8472/api/v1 \
#   HARBOR_SMOKE_EMAIL=you@example.com \
#   HARBOR_SMOKE_PASSWORD='your-password' \
#   ./scripts/smoke.sh
#
set -euo pipefail

API_URL="${HARBOR_API_URL:-http://localhost:8472/api/v1}"
EMAIL="${HARBOR_SMOKE_EMAIL:-}"
PASSWORD="${HARBOR_SMOKE_PASSWORD:-}"

if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
  echo "Set HARBOR_SMOKE_EMAIL and HARBOR_SMOKE_PASSWORD to run the smoke test." >&2
  exit 2
fi

# Build a fresh binary and run with an isolated, throwaway HOME.
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
make build >/dev/null
BIN="$ROOT/build/harbor"
HOME="$(mktemp -d)"; export HOME
trap 'rm -rf "$HOME"' EXIT

h() { "$BIN" --api-url "$API_URL" "$@"; }
step() { printf '\n\033[1;36m== %s ==\033[0m\n' "$*"; }

step "login as $EMAIL"
printf '%s\n' "$PASSWORD" | h login --email "$EMAIL"

step "whoami"
h whoami

step "create a notebook"
NB=$(h notebooks create --name "Smoke $(date +%s)" --json | jq -r '.id')
echo "notebook id: $NB"

step "create a Markdown note via stdin"
NOTE=$(printf '# Smoke Test\n\n- created by smoke.sh\n' | h notes create --title "Smoke note" --notebook "$NB" --stdin --json | jq -r '.note.id')
echo "note id: $NOTE"

step "tag the note (by name, auto-creates the tag)"
h notes tag "$NOTE" --tag-name smoke

step "list the note's tags"
h notes tags "$NOTE"

step "full-text search for the note"
h search "Smoke" --json | jq -r '.data[] | "\(.type)\t\(.title // .filename)"'

step "export the notebook to ENEX"
OUT="$HOME/export.enex"
h export enex --notebook "$NB" --output "$OUT"
echo "exported $(wc -c < "$OUT") bytes to $OUT"

step "trash the note, then restore it"
h notes delete "$NOTE"
h trash restore "$NOTE" >/dev/null && echo "restored"

step "logout"
h logout

printf '\n\033[1;32mSmoke test passed.\033[0m\n'
