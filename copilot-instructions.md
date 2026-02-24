# Marellis recipe site

This is a website to store and share recipes. The website is meant as a hobby project, built for my family to create and share recipes. The website is built using modern but simply web technologies (HTML5, JavaScript). The website will be hosted on GitHub Pages, and the source code is publicly available at GitHub.

## Project Structure

```
recipe-website-AI/
├── frontend/                # Frontend code (HTML, CSS, JS)
│   ├── index.html
│   ├── app.js
│   ├── api.js              # API client library
│   ├── style.css
│   └── README.md
├── backend/                # Backend code (Go API)
│   ├── cmd/api/main.go
│   ├── internal/
│   ├── migrations/
│   ├── go.mod
│   └── README.md
├── README.md              # Project overview
└── SETUP.md               # Setup & deployment guide
```

## Tech stack in use

### Backend

- Go is used for the API layer
  - Framework: Chi router (lightweight, idiomatic Go)
  - Database: SQLite with sqlx (single-file, JSON column support)
  - Architecture: Repository pattern with clean separation of concerns
  - Project structure: `backend/` with standard Go layout
  - API endpoints: RESTful API at `/api/v1/recipes` and `/api/v1/categories`
  - Deployment: Designed for Fly.io with Docker support
- SQLite is used as the only database component
  - JSON columns for arrays (categories, ingredients, steps, links)
  - Auto-increment integer IDs
  - Full-text search support

### Frontend

- HTML5, CSS and JavaScript are used for the front-end
- Static pages are hosted on GitHub Pages (from `frontend/` folder)
- `frontend/api.js` provides API client for backend communication
- Environment-aware API URL (localhost for dev, production URL for deployed)