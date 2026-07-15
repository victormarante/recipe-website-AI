# Deployment And CI/CD

The current deployment layout is:

- Backend API: Fly.io app `recipe-website-ai`
- Frontend: GitHub Pages from the `frontend/` directory
- Validation and deployment: GitHub Actions

## Backend Deployment

Fly.io configuration is at the repository root:

- `fly.toml`
- `Dockerfile`

Because `fly.toml` is root-level and has an empty `[build]` section, `flyctl deploy` uses the root Docker build context and the root `Dockerfile`. The separate `backend/Dockerfile` is not the canonical Fly deployment path in the current layout.

Required Fly secrets:

```bash
flyctl secrets set AUTH_USERNAME="..."
flyctl secrets set AUTH_PASSWORD="..."
flyctl secrets set JWT_SECRET="..."
flyctl secrets set CORS_ORIGIN="https://<owner>.github.io"
```

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
  - Uses `actions/setup-go@v5` with Go `1.21`
  - Runs `go mod download`, `gofmt` check, `go build ./...`, and `go test ./...`
  - Checks that key frontend files exist
- `.github/workflows/deploy-backend.yml`
  - Runs on every push to `master`
  - Deploys with `flyctl deploy --remote-only`
- `.github/workflows/deploy-frontend.yml`
  - Runs on every push to `master`
  - Deploys `frontend/` to GitHub Pages

Important current limitations:

- Deployment workflows are not gated on successful validation.
- Backend deployment is not path-filtered; every push to `master` triggers it.
- The validation workflow uses Go `1.21`, while `backend/go.mod` and the root Dockerfile use Go `1.24`.
- Some actions are pinned to moving tags or branches rather than commit SHAs.

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

- The backend currently starts with `http.ListenAndServe` and does not implement graceful shutdown.
- SQLite persistence depends on the Fly volume mounted at `/data`.
- Database backup, restore, readiness, and recovery procedures are not yet documented.
- There is no `/docs` API documentation endpoint in the application.

## Useful Commands

```bash
flyctl status
flyctl logs
flyctl releases
flyctl releases rollback
```
