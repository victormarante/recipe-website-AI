#!/usr/bin/env bash
# Imports recipes from a JSON array of CreateRecipeRequest objects into a running backend.
# Usage: backend/scripts/import-recipes.sh <recipes.json> [API_BASE_URL]   (default http://localhost:8080)
# Env:   ADMIN_PIN (default 123456, the development default)
# A title that already exists (in the backend or earlier in the file) gets a numeric suffix:
# "Title 2", "Title 3", ... Not idempotent: running it twice creates "... 2" copies of everything.
set -euo pipefail

FILE="${1:?usage: $0 <recipes.json> [API_BASE_URL]}"
BASE="${2:-http://localhost:8080}/api/v1"
PIN="${ADMIN_PIN:-123456}"

lower() { printf '%s' "$1" | jq -Rr ascii_downcase; }

TOKEN=$(curl -fsS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "$(jq -n --arg pin "$PIN" '{pin:$pin}')" | jq -r .token)

existing=$(curl -fsS "$BASE/recipes" | jq -r '.[]?.title | ascii_downcase')

while IFS= read -r recipe; do
  title=$(jq -r .title <<<"$recipe")
  candidate="$title"
  n=1
  while grep -qxF -- "$(lower "$candidate")" <<<"$existing"; do
    n=$((n + 1))
    candidate="$title $n"
  done

  curl -sS --fail-with-body -X POST "$BASE/recipes" \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d "$(jq --arg t "$candidate" '.title = $t' <<<"$recipe")" > /dev/null
  existing+=$'\n'"$(lower "$candidate")"
  echo "Skapade: $candidate"
done < <(jq -c '.[]' "$FILE")
