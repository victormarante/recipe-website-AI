# GitHub Copilot Instructions

- Follow the current repository structure: static frontend in `frontend/`, Go API in `backend/`.
- Keep frontend code vanilla JavaScript, HTML, and CSS. Do not suggest React, Vue, bundlers, or package managers unless explicitly requested.
- Keep backend code idiomatic Go using the existing Chi, sqlx, SQLite, and package layout.
- Treat all `/api/v1/recipes`, `/api/v1/categories`, and image routes as JWT-protected. Include bearer-token handling in generated examples.
- Do not suggest editing existing migrations; add numbered migrations for schema changes.
- Do not hard-code secrets or credentials. Use environment variables and Fly/GitHub secrets.
- Prefer small changes that match existing patterns over broad rewrites.
- Update documentation when changing endpoints, configuration, validation, deployment, or setup behavior.
