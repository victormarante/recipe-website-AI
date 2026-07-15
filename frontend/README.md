# Frontend

Static frontend for the Marellis Recipe Website, built with HTML, CSS, and vanilla JavaScript.

## Structure

```text
frontend/
├── index.html
├── app.js
├── api.js
├── style.css
└── README.md
```

## Current Behavior

- Shows a login overlay on first load
- Stores the JWT in `sessionStorage`
- Calls the backend API through `frontend/api.js`
- Supports browsing categories, searching, viewing details, and recipe CRUD
- Supports optional oven temperature
- Supports optional image upload/delete when backend R2 storage is configured
- Uses no frontend build tool or package manager

## API Configuration

`api.js` selects the backend URL by hostname:

- Local/non-GitHub Pages: `http://localhost:8080`
- GitHub Pages: `https://recipe-website-ai.fly.dev`

All recipe and category requests include `Authorization: Bearer <token>` when a token is present. Login is sent to `/api/v1/auth/login`.

## Run Locally

Start the backend first, then serve the static files:

```bash
cd frontend
python -m http.server 8000
```

Open `http://localhost:8000`.

## Manual Validation

There is no automated frontend test suite today. For frontend changes, check at minimum:

- The page loads without console errors
- Login works with configured backend credentials
- Recipes load after login
- Search and category filtering work
- Create, edit, and delete work
- Image upload/delete work when R2 is configured, and fail clearly when it is not
- Mobile and desktop layouts remain usable

## Deployment

GitHub Pages deployment uploads the `frontend/` directory with `.github/workflows/deploy-frontend.yml`.
