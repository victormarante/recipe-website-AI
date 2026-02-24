# Marellis Recipe Website

A family hobby project for storing and sharing recipes. Vanilla JavaScript frontend with a production-ready Go REST API backed by SQLite.

## Tech Stack

- **Frontend**: HTML5 · CSS3 · Vanilla JavaScript (ES6+)
- **Backend**: Go 1.21+ · Chi router · SQLite database
- **Deployment**: Docker · Fly.io

## Setup Guide

### Prerequisites

- **Go 1.21+** — [Download](https://golang.org/dl/) and install
- **Local web server** — Python, Node.js, or VS Code Live Server for frontend

### Install Go (Windows)

1. Download the `.msi` installer
2. Run installer and follow prompts
3. Verify in PowerShell: `go version`

### Quick Start

**Terminal 1 — Backend:**
```powershell
cd backend
go mod download
go run cmd/api/main.go
```
Server runs on `http://localhost:8080`

**Terminal 2 — Frontend:**
```powershell
cd frontend
python -m http.server 8000
```
Open `http://localhost:8000` in your browser

### Full Setup

#### 1. Backend Setup

```powershell
cd backend
```

**Download dependencies:**
```powershell
go mod download
go mod tidy
```

**Create environment file:**
```powershell
Copy-Item .env.example .env
```

Edit `.env`:
```env
PORT=8080
APP_ENV=development
DATABASE_PATH=./recipes.db
CORS_ORIGIN=http://localhost:8080,https://yourusername.github.io
```

**Run server:**
```powershell
go run cmd/api/main.go
# Or: make run (if make is installed)
```

**Test API:**
```powershell
curl http://localhost:8080/health
```

#### 2. Frontend Setup

```powershell
cd frontend
python -m http.server 8000
```

Navigate to `http://localhost:8000`

### API Testing Examples

**Get all recipes:**
```powershell
Invoke-RestMethod http://localhost:8080/api/v1/recipes
```

**Filter by category:**
```powershell
Invoke-RestMethod "http://localhost:8080/api/v1/recipes?category=breakfast"
```

**Search recipes:**
```powershell
Invoke-RestMethod "http://localhost:8080/api/v1/recipes?q=pancake"
```

**Create recipe:**
```powershell
$body = @{
    title = "Test Pancakes"
    description = "Fluffy pancakes"
    categories = @("breakfast")
    ingredients = @("1 cup flour", "2 eggs")
    steps = @("Mix", "Cook")
    links = @()
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/recipes" `
  -Method POST -Body $body -ContentType "application/json"
```

### Development Workflow

**Backend changes:**
1. Edit Go files
2. Restart server (Ctrl+C, then run again)

**Frontend changes:**
1. Edit HTML/CSS/JS
2. Refresh browser (no restart needed)

**Add migrations:**
1. Create `.sql` file in `migrations/`
2. Update `cmd/api/main.go`
3. Restart server

### Building for Production

**Build binary:**
```powershell
go build -o bin/api.exe cmd/api/main.go
.\bin\api.exe
```

**Build Docker image:**
```powershell
cd backend
docker build -t recipe-backend .
docker run -p 8080:8080 recipe-backend
```

### Deploy to Fly.io

**Prerequisites:**
- Fly CLI: [install guide](https://fly.io/docs/hands-on/install-flyctl/)
- Account: `flyctl auth signup`

**Deploy:**
```powershell
cd backend
flyctl launch      # First time only
flyctl deploy      # Deploy updates
flyctl logs        # View logs
flyctl open        # Open in browser
```

**After deployment:**
```powershell
# Update CORS for production
flyctl secrets set CORS_ORIGIN="https://yourusername.github.io"
```

Update `frontend/api.js`:
```javascript
const API_CONFIG = {
  development: 'http://localhost:8080',
  production: 'https://your-app.fly.dev', // Update this!
};
```

### Troubleshooting

| Issue | Solution |
|-------|----------|
| "go: command not found" | Install Go and restart terminal |
| Port 8080 already in use | Stop other app or change `PORT` in `.env` |
| CORS errors | Verify backend running + check `CORS_ORIGIN` in `.env` |
| Database locked | Only one server instance should run at a time |
| Dependencies not found | Run `go mod download && go mod tidy` |

## API Endpoints

```
GET    /health                    Health check
GET    /api/v1/recipes           Get all recipes (supports ?category= and ?q=)
POST   /api/v1/recipes           Create recipe
GET    /api/v1/recipes/:id       Get recipe by ID
PUT    /api/v1/recipes/:id       Update recipe
DELETE /api/v1/recipes/:id       Delete recipe
GET    /api/v1/categories        Get all categories
```

## Documentation

- [Backend API Docs](backend/README.md) — Full API reference
- [Frontend Docs](frontend/README.md) — Frontend architecture
- [Tech Instructions](copilot-instructions.md) — Project specifications

---

Made with ❤️ for family recipe sharing
