# Codex Repository Instructions

## Scope And Boundaries

- `frontend/` is a static HTML/CSS/vanilla JS application.
- `backend/` is a Go API using Chi, sqlx, and SQLite.
- Root `fly.toml` and root `Dockerfile` are the current Fly.io deployment path.
- Do not change runtime behavior, database schema, workflows, Dockerfiles, or deployment config during documentation-only tasks.
- Prefer one root `AGENTS.md`; do not add nested agent files unless the repo develops clearly different local rules.

## Inspect Before Substantial Changes

Read the relevant files first, and include these when the change affects their area:

- Root docs: `README.md`, `SETUP.md`, `DEPLOYMENT.md`, `TODO.md`
- Backend: `backend/go.mod`, `backend/.env.example`, `backend/cmd/api/main.go`, `backend/internal/router/router.go`, relevant handlers/repositories/models, and `backend/migrations/`
- Frontend: `frontend/index.html`, `frontend/app.js`, `frontend/api.js`, `frontend/style.css`
- Deployment: `.github/workflows/`, `fly.toml`, root `Dockerfile`, `backend/Dockerfile`

## Implementation Principles

- Keep changes small and aligned with existing patterns.
- Do not add dependencies or frameworks without explicit need and owner agreement.
- Keep backend handlers, repository code, and models consistent with the existing package layout.
- Keep frontend code framework-free and build-free unless explicitly requested.
- Do not document planned behavior as implemented.

## Backend Commands

Run from `backend/`:

```bash
gofmt -w ./...
go build ./...
go test ./...
go vet ./...
```

Use `make run`, `make test`, and `make test-coverage` only when helpful; the Makefile wraps standard Go commands.

## Frontend Validation

There is no frontend build step. For frontend changes:

- Verify `frontend/index.html`, `frontend/app.js`, and `frontend/api.js` still exist.
- Run a local static server when behavior changes.
- Manually check login, recipe CRUD, search/filtering, image behavior if relevant, and responsive layout.

## Database Migration Rules

- Do not edit existing migrations after they may have been applied.
- Add numbered migrations for schema changes.
- Avoid new ad hoc runtime schema changes in `cmd/api/main.go`.
- Preserve SQLite single-writer assumptions unless a task explicitly changes storage behavior.

## API Change Rules

- Treat `/api/v1` paths and JSON response fields as a contract.
- Keep authentication requirements explicit in docs and examples.
- Update backend docs and frontend API calls together when endpoint behavior changes.
- Distinguish public endpoints from JWT-protected endpoints.

## Security Constraints

- Never commit `.env`, secrets, tokens, database files, or private credentials.
- Keep `AUTH_USERNAME`, `AUTH_PASSWORD`, and `JWT_SECRET` required for backend startup.
- Do not expose internal error details in new API responses.
- Treat R2 credentials and Fly/GitHub tokens as secrets.

## Documentation Expectations

Update docs when behavior, setup, endpoints, configuration, validation, or deployment changes. Keep responsibilities separate:

- `README.md`: generic project overview and source of truth
- `backend/README.md`: backend API and operations
- `frontend/README.md`: frontend structure and validation
- `DEPLOYMENT.md`: deployment and CI/CD
- `TODO.md`: proposed future work, not completed work
- `CLAUDE.md` and `.github/copilot-instructions.md`: tool-specific guidance only

## Definition Of Done

- Requested changes are implemented and scoped.
- Formatting/build/test/validation checks relevant to the change were run.
- Documentation is updated when needed.
- No unrelated files were changed.
- Final report states which checks ran and which could not be run.
