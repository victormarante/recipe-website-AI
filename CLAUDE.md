# Claude Instructions

These instructions are for Claude-specific behavior in this repository. Generic project facts, architecture, setup, and API details live in `README.md`, `backend/README.md`, `frontend/README.md`, and `AGENTS.md`.

## Operating Mode

- Make the smallest safe change that solves the request.
- Read relevant files before editing.
- Prefer localized edits over broad refactors.
- Do not introduce new frameworks, dependencies, or deployment changes unless explicitly requested.
- For larger changes, state a short plan before editing.

## Response Style

- Be direct, technical, and concise.
- Prefer this shape for non-trivial work:
  - Plan: up to 3 bullets
  - Execution notes: only the important actions
  - Result: changed files and checks run
- Avoid long background explanations unless the user asks for them.

## Tool And Workflow Expectations

- Inspect the current implementation instead of relying on stale documentation.
- Preserve user changes in the working tree.
- Do not rewrite history, force-push, delete branches, or make destructive git changes without explicit instruction.
- Ask before pushing or opening a PR unless the user has clearly requested that workflow.

## Verification Commands

- Backend (from `backend/`): `gofmt -l ./...` (format check), `go build ./...`, `go vet ./...`, `go test ./...`.
- `go` may not be on PATH in some sandboxes — if `go` is not found, try `/usr/local/go/bin/go`.
- `make lint` in `backend/` calls `golangci-lint`, which is not installed in this environment — don't rely on it; `go vet` is the available substitute.
- Frontend has no package manager, build step, or formatter (no `package.json` anywhere). Validate with `node --check frontend/app.js` and `node --check frontend/api.js`; match existing style by hand.

## Project Constraints

- Backend API has no authentication (removed deliberately) — don't reintroduce auth unless explicitly asked.
- `/api/v1` is a stable contract — don't change existing endpoint request/response shapes without explicit instruction.
- Never edit applied migration files in `backend/migrations/`; add new numbered migrations instead.
- Backend deploys via the root `Dockerfile` + `fly.toml` — there is no separate backend-only Dockerfile.

## Reporting

At the end of implementation work, report:

- What changed
- Which checks were run
- Which checks could not be run, if any
- Any follow-up risk that needs owner review
