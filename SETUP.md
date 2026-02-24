# Setup Guide: Marellis Recipe Backend

This guide walks you through setting up and running the Go backend API for your recipe website.

## Prerequisites

You need to install Go before running the backend. The code has been created but requires Go to execute.

### Install Go

1. Download Go from [https://golang.org/dl/](https://golang.org/dl/)
2. Choose the Windows installer (`.msi` file)
3. Run the installer and follow the prompts
4. Verify installation by opening a new PowerShell and running:
   ```powershell
   go version
   ```
   You should see something like `go version go1.21.x windows/amd64`

## Quick Start

Once Go is installed, follow these steps:

### 1. Navigate to Backend Directory

```powershell
cd backend
```

### 2. Download Dependencies

```powershell
go mod download
go mod tidy
```

This will install all required packages:
- Chi router
- SQLite driver
- CORS middleware
- Validator
- And more...

### 3. Create Environment File

```powershell
Copy-Item .env.example .env
```

Edit `.env` and update the `CORS_ORIGIN` if needed:
```env
PORT=8080
APP_ENV=development
DATABASE_PATH=./recipes.db
CORS_ORIGIN=http://localhost:8080,https://yourusername.github.io
```

### 4. Run the Server

```powershell
go run cmd/api/main.go
```

Or using the Makefile (if you have `make` installed):
```powershell
make run
```

You should see output like:
```
Starting Recipe API Server in development mode...
Database connected: ./recipes.db
Database migrations completed
Server listening on port 8080
CORS origins: [http://localhost:8080 https://yourusername.github.io]
API endpoints:
  GET    /health
  GET    /api/v1/recipes
  POST   /api/v1/recipes
  ...
```

### 5. Test the API

Open a new PowerShell terminal and test the health endpoint:

```powershell
curl http://localhost:8080/health
```

You should see: `OK`

Create a test recipe:

```powershell
$body = @{
    title = "Test Pancakes"
    description = "Fluffy test pancakes"
    categories = @("breakfast", "test")
    ingredients = @("1 cup flour", "2 eggs")
    steps = @("Mix ingredients", "Cook")
    links = @()
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/recipes" -Method POST -Body $body -ContentType "application/json"
```

Get all recipes:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/recipes"
```

## Running the Full Application

### Backend (Terminal 1)

```powershell
cd backend
go run cmd/api/main.go
```

### Frontend (Terminal 2)

You need a local web server to avoid CORS issues. You can use Python's built-in server:

```powershell
# Navigate to frontend folder
cd frontend

# If you have Python 3 installed:
python -m http.server 8000

# Or use VS Code Live Server extension
# Right-click frontend/index.html and select "Open with Live Server"
```

Then open your browser to:
- Frontend: `http://localhost:8000` (or whichever port your server uses)
- The frontend will automatically connect to the backend at `http://localhost:8080/api/v1`

## Development Workflow

### Making Changes to Backend

1. Edit Go files in `backend/`
2. Stop the server (Ctrl+C)
3. Restart with `go run cmd/api/main.go`
4. Test changes

### Making Changes to Frontend

1. Edit files in `frontend/` folder (`index.html`, `app.js`, `style.css`, or `api.js`)
2. Refresh browser (no restart needed)

### Adding Database Migrations

1. Create new `.sql` file in `migrations/` directory
2. Update `cmd/api/main.go` to run the new migration
3. Restart server

## Testing

### Manual API Testing

Use PowerShell `Invoke-RestMethod` or install a tool like:
- [Postman](https://www.postman.com/downloads/)
- [Insomnia](https://insomnia.rest/download)
- [HTTPie](https://httpie.io/cli)

### Example Requests

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

**Get categories:**
```powershell
Invoke-RestMethod http://localhost:8080/api/v1/categories
```

## Building for Production

### Build Binary

```powershell
go build -o bin/api.exe cmd/api/main.go
```

Run the binary:
```powershell
.\bin\api.exe
```

### Build Docker Image

```powershell
cd backend
docker build -t recipe-backend .
docker run -p 8080:8080 recipe-backend
```

## Deployment to Fly.io

### Prerequisites

1. Install Fly CLI: [https://fly.io/docs/hands-on/install-flyctl/](https://fly.io/docs/hands-on/install-flyctl/)
2. Create account: `flyctl auth signup`
3. Log in: `flyctl auth login`

### Deploy

```powershell
# Navigate to backend folder
cd backend

# Initialize Fly app (first time only)
flyctl launch

# Deploy updates
flyctl deploy

# View logs
flyctl logs

# Check status
flyctl status

# Open in browser
flyctl open
```

### Important Deployment Notes

1. **Update CORS Origin**: After deploying, update your `.env` or Fly secrets with your production GitHub Pages URL
   ```powershell
   flyctl secrets set CORS_ORIGIN="https://yourusername.github.io"
   ```

2. **Persistent Volume**: SQLite needs persistent storage. Fly.io will configure this during `flyctl launch`

3. **Update Frontend**: Update `frontend/api.js` production URL to your Fly.io app URL:
   ```javascript
   const API_CONFIG = {
     development: 'http://localhost:8080',
     production: 'https://your-app.fly.dev', // Update this!
   };
   ```

## Troubleshooting

### "go: command not found"

Go is not installed or not in your PATH. Install Go and restart your terminal.

### Database locked errors

SQLite allows only one writer at a time. This is normal and the code handles it. If issues persist, check that only one instance of the server is running.

### CORS errors in browser

1. Check that backend is running
2. Verify `CORS_ORIGIN` in `.env` matches your frontend URL
3. Check browser console for specific error

### Port already in use

Another application is using port 8080. Either:
- Stop the other application
- Change `PORT` in `.env` to a different port (e.g., 8081)

### Dependencies not found

Run `go mod download` and `go mod tidy` to install dependencies.

## Next Steps

1. **Install Go** if you haven't already
2. **Run the backend** following the Quick Start section
3. **Test with curl/Postman** to verify API works
4. **Run the frontend** with a local web server
5. **Migrate localStorage data** (if you have existing recipes)
6. **Deploy to production** when ready

## Project Structure Reference

```
recipe-website-AI/
├── backend/
│   ├── cmd/api/main.go              # Entry point
│   ├── internal/
│   │   ├── config/config.go         # Environment configuration
│   │   ├── database/db.go           # Database connection
│   │   ├── handlers/
│   │   │   ├── recipes.go           # Recipe endpoints
│   │   │   └── categories.go        # Category endpoints
│   │   ├── middleware/logger.go     # Request logging
│   │   ├── models/recipe.go         # Data models
│   │   ├── repository/              # Database layer
│   │   │   └── recipe_repository.go
│   │   └── router/router.go         # Route definitions
│   ├── migrations/
│   │   └── 001_create_tables.sql    # Database schema
│   ├── .env.example                 # Environment template
│   ├── Dockerfile                   # Container build
│   ├── Makefile                     # Build commands
│   └── README.md                    # API documentation
├── frontend/
│   ├── index.html
│   ├── app.js
│   ├── api.js                       # API client
│   ├── style.css
│   └── README.md                    # Frontend documentation
├── README.md                        # Project overview
└── SETUP.md                         # This file
```

## Support

If you encounter issues:
1. Check the [Go documentation](https://golang.org/doc/)
2. Review [Chi router docs](https://go-chi.io/)
3. Check [SQLite Go driver docs](https://pkg.go.dev/modernc.org/sqlite)

Happy coding! 🍽️
