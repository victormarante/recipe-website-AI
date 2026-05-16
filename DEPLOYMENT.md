# Deployment & CI/CD Guide

## 📋 Overview

This project uses GitHub Actions for automated deployment:

- **Validation Workflow** - Runs on every push and pull request
- **Backend Deployment** - Automatically deploys to Fly.io when backend changes
- **Frontend Deployment** - Automatically deploys to GitHub Pages when frontend changes

---

## 🔧 Prerequisites

### 1. GitHub Repository Secrets

Set up the following secrets in your GitHub repository settings (`Settings → Secrets and variables → Actions`):

#### Required Secrets

**`FLY_API_TOKEN`** - Authentication token for Fly.io deployment

```bash
# Generate token via Fly.io:
flyctl auth token
```

1. Get the token from `flyctl auth token`
2. Go to GitHub repo → Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `FLY_API_TOKEN`
5. Value: Paste the token from step 1

**`GITHUB_TOKEN`** - Automatically provided by GitHub

- No manual setup needed; GitHub automatically provides this
- Used for GitHub Pages deployment

### 2. Fly.io Setup

Ensure you've initialized your Fly.io app:

```bash
cd backend
flyctl launch  # First time only
```

This creates `fly.toml` with your app configuration.

### 3. GitHub Pages Configuration

Enable GitHub Pages in your repository:

1. Go to `Settings → Pages`
2. Source: "Deploy from a branch"
3. Branch: `master`
4. Folder: `/ (root)` (we'll use workflows to deploy from `/frontend`)

---

## 🚀 Deployment Workflows

### Workflow 1: Validation (`validate.yml`)

**Triggers:** Every push and pull request

**Jobs:**
1. **Validate Backend**
   - Check Go formatting
   - Run tests
   - Upload code coverage

2. **Validate Frontend**
   - Verify HTML files
   - Check JavaScript syntax
   - Validate CSS

3. **Security Check**
   - Scan for exposed secrets
   - Check `.gitignore` configuration

**Status:** Visible on PR and commits ✅/❌

### Workflow 2: Backend Deployment (`deploy-backend.yml`)

**Triggers:** Push to `master` branch with changes in `backend/**`

**Steps:**
1. Checkout code
2. Setup Go 1.21
3. Download dependencies (`go mod download`)
4. Run tests (`go test`)
5. Deploy to Fly.io using `flyctl deploy`

**Deployment Target:** `recipe-website-ai.fly.dev`

**Status:** Check Actions tab for logs

### Workflow 3: Frontend Deployment (`deploy-frontend.yml`)

**Triggers:** Push to `master` branch with changes in `frontend/**`

**Steps:**
1. Checkout code
2. Validate frontend files exist
3. Deploy to GitHub Pages using `peaceiris/actions-gh-pages@v3`

**Deployment Target:** `https://yourusername.github.io/recipe-website-AI/`

**Status:** Check Actions tab for logs

---

## 📝 Making Changes & Deploying

### Scenario 1: Update Backend Code

```bash
# Make changes to backend files
nano backend/internal/handlers/recipes.go

# Commit and push
git add backend/
git commit -m "Add new recipe filtering"
git push origin master
```

**Automatic Actions:**
1. Validation workflow runs ✅
2. Backend deployment starts
3. Code deployed to Fly.io
4. New version live at `recipe-website-ai.fly.dev`

### Scenario 2: Update Frontend Code

```bash
# Make changes to frontend files
nano frontend/app.js

# Commit and push
git add frontend/
git commit -m "Improve UI responsiveness"
git push origin master
```

**Automatic Actions:**
1. Validation workflow runs ✅
2. Frontend deployment starts
3. Code deployed to GitHub Pages
4. New version live at `yourusername.github.io/recipe-website-AI`

### Scenario 3: Update Both

```bash
git add backend/ frontend/
git commit -m "Release v1.1: Enhanced security & UI"
git push origin master
```

**Automatic Actions:**
1. Validation workflow runs ✅
2. Backend deployment starts → Fly.io
3. Frontend deployment starts → GitHub Pages
4. Both deployed simultaneously

---

## 🔍 Monitoring Deployments

### View Workflow Status

1. Go to GitHub repo → **Actions** tab
2. See all workflow runs
3. Click on a workflow to see details

### Workflow Run Details

- **Status:** ✅ Success, ❌ Failed, ⏳ In Progress
- **Duration:** How long the workflow took
- **Logs:** Detailed output from each step
- **Artifacts:** Any generated files (e.g., code coverage)

### Troubleshooting Failed Deployments

1. Click on the failed workflow run
2. Expand the failed job
3. Check the error message
4. Common issues:
   - Missing secrets (FLY_API_TOKEN)
   - Tests failing
   - Code formatting issues
   - File not found

### Viewing Recent Deployments

**Fly.io:**
```bash
flyctl deployed
flyctl logs
```

**GitHub Pages:**
- Check repo → Actions → "Deploy Frontend to GitHub Pages"
- Deployments tab shows all releases

---

## 🔐 Secrets Management

### Setting Secrets

1. Go to repo → **Settings**
2. **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Enter `Name` and `Value`
5. Click **Add secret**

### Using Secrets in Workflows

Secrets are referenced with `${{ secrets.SECRET_NAME }}`

```yaml
- name: Deploy to Fly.io
  env:
    FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
  run: flyctl deploy
```

### Sensitive Data

Never commit:
- `.env` files
- API tokens
- Database credentials
- JWT secrets

Use `.gitignore` to prevent accidental commits:

```
.env
.env.local
*.key
secrets/
```

---

## 🛠️ Customizing Workflows

### Change Trigger Branch

Edit workflow files to trigger on different branches:

```yaml
on:
  push:
    branches:
      - develop  # Change from master to develop
```

### Add Notifications

Add Slack or email notifications:

```yaml
- name: Notify Slack
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### Schedule Deployments

Deploy on a schedule (e.g., nightly):

```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM UTC daily
```

---

## 📊 Status Badges

Add deployment status to your README:

```markdown
## Status

![Validation](https://github.com/victormarante/recipe-website-AI/workflows/Validate%20&%20Test%20Code/badge.svg)
![Backend Deploy](https://github.com/victormarante/recipe-website-AI/workflows/Deploy%20Backend%20to%20Fly.io/badge.svg)
![Frontend Deploy](https://github.com/victormarante/recipe-website-AI/workflows/Deploy%20Frontend%20to%20GitHub%20Pages/badge.svg)
```

---

## 📝 Production Checklist

Before deploying to production:

- [ ] Set `FLY_API_TOKEN` secret in GitHub
- [ ] Update `.env` secrets in Fly.io:
  ```bash
  flyctl secrets set JWT_SECRET="strong-random-secret-32-chars-min"
  flyctl secrets set AUTH_USERNAME="secure-username"
  flyctl secrets set AUTH_PASSWORD="secure-password"
  flyctl secrets set CORS_ORIGIN="https://yourusername.github.io/recipe-website-AI"
  ```
- [ ] Enable HTTPS on Fly.io (set in `fly.toml`)
- [ ] Update frontend `api.js` with production API URL
- [ ] Test login with production credentials
- [ ] Verify API documentation accessible at `/docs`
- [ ] Monitor logs: `flyctl logs`

---

## 🔄 Deployment Rollback

### If Deployment Fails

**Fly.io:**
```bash
# View deployment history
flyctl releases

# Rollback to previous version
flyctl releases rollback
```

**GitHub Pages:**
- GitHub Pages keeps previous versions
- Revert commit and push to redeploy
```bash
git revert <commit-hash>
git push origin master
```

---

## 📈 Performance Monitoring

### Fly.io Metrics

```bash
flyctl status
flyctl metrics
```

### GitHub Pages

- Deployment time usually < 1 minute
- Check Actions for any slow builds

---

## 🆘 Common Issues

### Workflow Doesn't Run

**Issue:** Push to master but workflow doesn't start

**Solutions:**
- Check branch name is exactly `master`
- Verify files changed match path filters
- Check Actions are enabled in repo settings
- Verify YAML syntax in workflow file

### FLY_API_TOKEN Error

**Issue:** `Error: No api token found`

**Solution:**
1. Go to repo → Settings → Secrets
2. Verify `FLY_API_TOKEN` exists
3. Regenerate token: `flyctl auth token`
4. Update secret with new token

### Frontend Not Deploying

**Issue:** Frontend changes not appearing on GitHub Pages

**Solutions:**
- Clear browser cache (Ctrl+Shift+R)
- Check Actions tab for deployment status
- Verify GitHub Pages settings point to correct branch
- Check `peaceiris/actions-gh-pages@v3` configuration

### Tests Failing

**Issue:** Validation workflow fails on tests

**Solutions:**
```bash
# Run tests locally first
cd backend
go test -v ./...

# Fix failing tests before pushing
```

---

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Fly.io Deployment Guide](https://fly.io/docs/getting-started/getting-started-with-golang/)
- [GitHub Pages Deployment Action](https://github.com/peaceiris/actions-gh-pages)
- [Setting GitHub Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

---

## 🎯 Next Steps

1. ✅ Set `FLY_API_TOKEN` secret
2. ✅ Push changes to trigger deployment
3. ✅ Monitor workflows in Actions tab
4. ✅ Verify backend at `recipe-website-ai.fly.dev`
5. ✅ Verify frontend at GitHub Pages URL
6. ✅ Test login and API functionality

Congratulations! Your application now has automated, secure deployments! 🎉
