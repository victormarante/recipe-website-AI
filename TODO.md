# TODO

This is a reviewable roadmap of proposed future work. Items are not commitments and should not be marked complete until implemented and verified.

## Current Work: Images, Routing, Access Control

### Image display
- [x] P1 Fix cropped recipe images in the detail view (`.recipe-photo`, e.g. recipe ID 2).

### Hash routing
- [x] P1 Add route parser and `hashchange` handling (`#/`, `#/recipes`, `#/recipes/<id>`, category list).
- [x] P1 Replace `pushState`/`popstate` navigation with hash navigation.
- [x] P1 Support deep links to a recipe (fetch by ID when not yet loaded; unknown ID falls back to list).

### Access control (admin vs regular)
- [x] P0 Backend: `ADMIN_PIN` (at least 6 digits) and `AUTH_SECRET` config, with tests.
- [x] P0 Backend: `POST /api/v1/auth/login`, signed token, `RequireAdmin` middleware on write routes only, login rate limiting.
- [x] P1 Backend tests: middleware, login handler, protected vs public routes.
- [x] P0 Frontend: token storage, `Authorization` header (including image upload), 401 handling.
- [x] P1 Frontend: login/logout UI and hide add/edit/delete controls in regular mode (Swedish strings).
- [x] P1 Docs: update CLAUDE.md, AGENTS.md, README, SETUP, DEPLOYMENT, backend README, `.env.example`.
- [x] P0 Owner: set `ADMIN_PIN` and `AUTH_SECRET` as Fly secrets before deploying.

## Documentation Follow-ups

- [x] P0 Fix authentication examples in API documentation (rewritten for PIN auth).
- [x] P1 Document image upload and delete endpoints in the backend API documentation.
- [x] P1 Make `backend/.env.example` match all configuration read by the application, including optional R2 variables.
- [x] P1 Add database readiness, backup, restore, and recovery documentation for the Fly.io SQLite volume.
- [x] P1 Document PIN authentication, token lifetime, and browser `localStorage` token storage (see Access Control).
- [ ] P2 Add a short manual QA checklist for the deployed GitHub Pages and Fly.io environments.

## Critical Correctness And Security

- [x] P0 Stop exposing internal error details in API responses through the `message` field.
- [x] P0 Distinguish not-found errors from internal server errors for update and delete paths.
- [x] P1 Improve image type validation beyond `http.DetectContentType` prefix checks, and define allowed formats explicitly.
- [x] P1 Improve image cleanup behavior when object upload succeeds but saving the URL fails.
- [x] P1 Add graceful shutdown for the HTTP server, or keep documentation clear that it is not implemented.

## Testing

- [x] P0 Add substantive repository tests using isolated SQLite databases.
- [x] P0 Add handler tests for auth, recipe CRUD, category responses, and image error paths.
- [x] P1 Add authentication middleware tests for missing, malformed, expired, and tampered tokens (code was removed; re-add with new auth).
- [x] P1 Add router tests that verify public versus protected routes.
- [ ] P2 Add lightweight frontend smoke tests or documented browser test scripts.

## Database And Migrations

- [x] P0 Replace ad hoc runtime schema changes for `oven_temperature` and `image_url` with numbered migrations.
- [x] P1 Decide how applied migrations are tracked before adding more schema changes.
- [x] P1 Add migration tests or startup checks that catch schema drift.
- [ ] P2 Review whether category filtering should normalize category values instead of relying on JSON text matching.

## API Consistency

- [ ] P2 Keep image upload and delete endpoint documentation aligned with future R2 behavior changes.
- [x] P1 Replace manual category JSON construction with standard JSON encoding.
- [x] P1 Make error response formats consistent between middleware, category handlers, and recipe handlers.
- [ ] P2 Decide whether links and optional fields should have stricter validation rules.

## Image Storage

- [x] P1 Add R2 variables to `.env.example` with safe placeholders and comments.
- [x] P1 Validate that `R2_BUCKET_NAME` and `R2_PUBLIC_URL` are present before enabling image endpoints.
- [x] P1 Decide and document object key strategy, overwrite behavior, and cache behavior for recipe images.
- [ ] P2 Add tests for missing R2 configuration and oversized/non-image uploads.

## CI And Deployment

- [x] P0 Standardize the Go version across `backend/go.mod`, retained Dockerfile, GitHub Actions, and documentation.
- [x] P0 Ensure deployment happens only after successful validation.
- [x] P1 Decide on and document one canonical Dockerfile; remove or clearly deprecate the other.
- [x] P1 Add `go vet` and stronger frontend validation to CI.
- [x] P1 Pin third-party GitHub Actions to stable versions or commit SHAs.
- [ ] P2 Add path filters or workflow dependencies if deployments should only run for relevant changes.

## Maintainability

- [ ] P1 Replace duplicated endpoint lists with one maintained API reference or clear ownership between README files. Deferred because this pass kept endpoint docs accurate while avoiding a larger documentation restructuring.
- [x] P1 Introduce typed sentinel errors in the repository layer so handlers can map errors reliably.
- [ ] P2 Sort category responses for stable frontend display and test output.
- [ ] P2 Review `backend/Makefile` targets and docs for commands that assume unavailable tools such as `golangci-lint`.

## Optional Future Improvements

- [ ] P2 Add SQLite FTS if real full-text search is desired.
- [ ] P2 Add structured request logging with request IDs.
- [ ] P2 Add an admin-friendly export/import flow for recipes and images.
