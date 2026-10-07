#!/usr/bin/env bash
# Starts the backend (:8080) and frontend (:8000) for local development.
# Usage: ./dev.sh        Ctrl+C stops both.
set -euo pipefail
cd "$(dirname "$0")"

GO=$(command -v go || echo /usr/local/go/bin/go)

[ -f backend/.env ] || cp backend/.env.example backend/.env

trap 'kill 0' EXIT INT TERM

(cd backend && "$GO" run cmd/api/main.go) &
(cd frontend && python3 -m http.server 8000) &

echo "Frontend: http://localhost:8000   Backend: http://localhost:8080"
wait
