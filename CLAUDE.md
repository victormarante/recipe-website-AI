# CLAUDE.md

## Project Overview

Marellis Recipe Website is a personal hobby project for storing and sharing recipes.

Architecture:
- Frontend: Vanilla JavaScript (HTML/CSS/JS)
- Backend: Go 1.21+ (Chi router)
- Database: SQLite
- Deployment: Fly.io (backend) + GitHub Pages (frontend)
- CI/CD: GitHub Actions

This is a production-like hobby system with real deployments and automation.

---

## Claude Operating Mode

You are working inside a production hobby project.

Primary goal: **make the smallest safe change that solves the problem.**

Optimization rules:
- Prefer minimal diffs
- Avoid unnecessary refactors
- Preserve existing architecture unless explicitly requested
- Do not introduce new frameworks or dependencies without instruction

Before making changes:
1. Read relevant files first
2. Identify minimal affected scope
3. Propose smallest viable change

---

## Communication Mode (Caveman Mode)

Default output style:
- short
- direct
- technical
- low-token

Format:
- Plan (max 3 bullets)
- Execution steps
- Result summary (max 3 bullets)

Avoid:
- long explanations
- repetition
- unnecessary context

---

## Project Map

### Backend (Go)
- `backend/cmd/api/main.go` → application entrypoint
- `backend/internal/handlers/` → HTTP handlers (recipes, categories, auth)
- `backend/internal/middleware/` → logger + JWT auth middleware
- `backend/internal/database/` → SQLite access layer
- `backend/internal/repository/` → data access layer
- `backend/internal/models/` → domain models
- `backend/internal/router/` → route definitions
- `backend/internal/config/` → environment config loader
- `backend/migrations/` → database schema changes
- `backend/.env` → local environment config (copy from `.env.example`)

### Frontend (Vanilla JS)
- `frontend/index.html` → main UI
- `frontend/app.js` → UI logic
- `frontend/api.js` → API client layer

### Deployment
- `fly.toml` → Fly.io configuration (at repo root)
- `Dockerfile` → backend Docker build (at repo root, used by Fly.io)
- `.github/workflows/validate.yml` → validation pipeline
- `.github/workflows/deploy-backend.yml` → backend deployment
- `.github/workflows/deploy-frontend.yml` → frontend deployment

---

## Data Model (SQLite)

### recipes

- id (INTEGER PRIMARY KEY AUTOINCREMENT)
- title (text)
- description (text)
- categories (JSON array of strings)
- ingredients (JSON array of strings)
- steps (JSON array of strings)
- links (JSON array of link objects: `{type, url, label, linked_recipe_id}`)
- created_at (timestamp)
- updated_at (timestamp, auto-updated via trigger)

Notes:
- SQLite is single-instance only
- Do not assume multi-writer concurrency
- Keep schema simple and explicit

---

## API Contract Rules

Base URL:
```
/api/v1
```

Endpoints:
- POST /api/v1/auth/login (public — returns JWT)
- GET /api/v1/recipes (protected)
- POST /api/v1/recipes (protected)
- GET /api/v1/recipes/:id (protected)
- PUT /api/v1/recipes/:id (protected)
- DELETE /api/v1/recipes/:id (protected)
- GET /api/v1/categories (protected)
- GET /health (public)

Rules:
- All responses are JSON
- Errors use:
  { "error": "message" }

Do NOT:
- change endpoint paths without explicit instruction
- rename response fields without instruction
- change HTTP methods without instruction

---

## Backend Rules (Go)

- Keep Go code idiomatic and simple
- Prefer standard library where possible
- Keep handlers thin, logic in service/db layers
- Avoid introducing ORMs or heavy frameworks
- SQLite queries should remain readable and explicit

---

## Frontend Rules

- Vanilla JS only (no frameworks)
- Keep UI logic simple and imperative
- Avoid adding build tools unless requested
- Do not introduce React/Vue/Svelte

---

## Deployment Rules

### Backend (Fly.io)
- Deploy via GitHub Actions or `flyctl deploy`
- Config in `fly.toml`
- Environment variables set via Fly secrets

### Frontend (GitHub Pages)
- Deployed automatically from GitHub Actions
- Do not manually modify production build output

### Environment Variables
Never commit secrets:
- `.env`
- API keys
- Fly tokens
- database credentials

Use:
- GitHub Secrets
- Fly.io secrets

---

## Safe Editing Rules

When modifying code:
- Prefer editing existing files over creating new ones
- Avoid duplicating logic
- Keep changes localized
- Do not restructure project unless explicitly asked

Before large changes:
- warn user
- propose plan

---

## Git Workflow Assumptions

- Main branch: `master`
- Workflows trigger on push to `master`

Before significant changes:
- recommend a checkpoint commit

Commit style:
- small
- descriptive
- incremental

Never:
- force push
- rewrite history
- delete branches without instruction

---

## Development Workflow

Backend:
```bash
cd backend
cp .env.example .env   # then set AUTH_USERNAME, AUTH_PASSWORD, JWT_SECRET
go run cmd/api/main.go
```

Frontend:
```bash
cd frontend
python -m http.server 8000
```

API testing:
- curl or PowerShell Invoke-RestMethod

---

## Testing & Validation

Before pushing:
- ensure backend builds
- ensure API endpoints respond
- ensure frontend loads without errors

CI runs:
- Go formatting checks
- backend tests
- frontend validation
- security checks

---

## Troubleshooting Awareness

Common issues:
- Go not installed → install Go 1.21+
- port conflicts → change PORT in .env
- CORS errors → verify CORS_ORIGIN
- DB locked → only one backend instance allowed
- Fly deploy issues → check FLY_API_TOKEN

---

## Scope Boundaries

You only operate inside this repository.

Do NOT:
- access work-related code
- access external projects
- modify system files
- assume cross-repo context

---

## Mental Model

Think:
- “minimal safe patch”
- “local change only”
- “no architectural drift unless requested”

---

## Default Priorities

1. Correctness
2. Simplicity
3. Maintainability
4. Performance (only if needed)
5. Cleverness (avoid unless necessary)
