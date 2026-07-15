# Deployment And CI/CD

The current deployment layout is:

- Backend API: Fly.io app `recipe-website-ai`
- Frontend: GitHub Pages from the `frontend/` directory
- Validation and deployment: GitHub Actions

## Backend Deployment

Fly.io configuration is at the repository root:

- `fly.toml`
- `Dockerfile`

Because `fly.toml` is root-level and has an empty `[build]` section, `flyctl deploy` uses the root Docker build context and the root `Dockerfile`. There is no second backend Dockerfile.

Required Fly secrets:

```bash
flyctl secrets set AUTH_USERNAME="..."
flyctl secrets set AUTH_PASSWORD_HASH='...'
flyctl secrets set JWT_SECRET="..."
flyctl secrets set CORS_ORIGIN="https://<owner>.github.io"
```

`AUTH_PASSWORD` is still supported for backwards compatibility. Prefer `AUTH_PASSWORD_HASH` for production. Generate a bcrypt hash locally with a trusted tool and do not commit it.

Optional image storage secrets:

```bash
flyctl secrets set R2_ACCOUNT_ID="..."
flyctl secrets set R2_ACCESS_KEY_ID="..."
flyctl secrets set R2_SECRET_ACCESS_KEY="..."
flyctl secrets set R2_BUCKET_NAME="..."
flyctl secrets set R2_PUBLIC_URL="..."
```

`DATABASE_PATH` is set in `fly.toml` to `/data/recipes.db`, and the app mounts the `recipe_data` volume at `/data`.

Deploy manually:

```bash
flyctl deploy --remote-only
flyctl logs
flyctl status
```

## Frontend Deployment

`.github/workflows/deploy-frontend.yml` uploads `frontend/` as a GitHub Pages artifact and deploys it with the official Pages actions.

The frontend production API base URL is currently hard-coded in `frontend/api.js` as:

```text
https://recipe-website-ai.fly.dev
```

Update that code only in a frontend change, not in documentation-only work.

## GitHub Actions

Current workflows:

- `.github/workflows/validate.yml`
  - Runs on pushes and pull requests to `master`
  - Uses `actions/setup-go@v5` with `go-version-file: backend/go.mod`
  - Runs `go mod download`, `gofmt` check, `go build ./...`, `go vet ./...`, and `go test ./...`
  - Checks that key frontend files exist and runs JavaScript/HTML syntax checks
- `.github/workflows/deploy-backend.yml`
  - Runs after the `Validate` workflow succeeds on `master`
  - Deploys with `flyctl deploy --remote-only`
- `.github/workflows/deploy-frontend.yml`
  - Runs after the `Validate` workflow succeeds on `master`
  - Deploys `frontend/` to GitHub Pages

Current notes:

- Deployment workflows are gated on successful validation, but are not path-filtered.
- The Fly setup action is pinned to tag `1.6`; official GitHub actions use stable major version tags.

These are tracked in `TODO.md`.

## Required GitHub Secret

Set this repository secret for backend deployment:

```text
FLY_API_TOKEN
```

Generate it with:

```bash
flyctl auth token
```

GitHub provides `GITHUB_TOKEN` automatically for Pages deployment.

## Operational Notes

- The backend uses `http.Server` with read/write/idle timeouts and graceful shutdown on `SIGINT`/`SIGTERM`.
- SQLite persistence depends on the Fly volume mounted at `/data`.
- `/health` is a liveness endpoint. `/ready` checks SQLite connectivity and is configured as the Fly health check.
- There is no `/docs` API documentation endpoint in the application.

## Database Backup And Restore

SQLite data lives at `/data/recipes.db` on the Fly volume. Before risky changes, take a Fly volume snapshot or copy the database file from a stopped or quiet app instance. Avoid copying during active writes.

Basic owner-run options:

```bash
flyctl volumes snapshots list
flyctl ssh console
```

Inside the machine, inspect `/data/recipes.db`. Restore procedures should be rehearsed with a non-production copy before relying on them in production.

## Useful Commands

```bash
flyctl status
flyctl logs
flyctl releases
flyctl releases rollback
```
