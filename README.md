# Postcube

Postcube is a BasaltPass-integrated anonymous question box app.
Each user gets a public page where visitors can ask anonymous questions, while the owner manages replies from a private inbox.

## Stack

- Backend: Go, Fiber, GORM, SQLite
- Frontend: React, TypeScript, Vite, Tailwind CSS v4
- Auth: BasaltPass OAuth2 PKCE + HttpOnly session cookie
- Default backend port: `8113`
- Default frontend port: `5116`

## Core Features

- Public question box per user: `/u/:slug`
- Anonymous submission form
- Public feed shows both answered and pending questions
- Private inbox for owner:
  - answer or clear answer (back to pending)
  - choose question background color
  - delete question
- Box title editing and sharable public URL

## Project Layout

```text
Postcube/
  backend/
    database/
    handlers/
    models/
    main.go
  frontend/
    src/
      api/
      hooks/
      pages/
  .basalt/
    app.json
    rbac.json
    resources.json
  register_postcube.py
  project.meta.json
```

## Backend Setup

```powershell
cd backend
cp .env.example .env
# fill BASALT_CLIENT_ID / BASALT_CLIENT_SECRET / JWT_SECRET

go mod tidy
go run .
```

Health check:

```text
http://localhost:8113/api/health
```

## Frontend Setup

```powershell
cd frontend
npm install
npm run dev
```

Open:

```text
http://localhost:5116
```

## Register App In BasaltPass

Use the helper script (requires your tenant API key):

```powershell
cd Postcube
$env:BASALT_API_KEY="your_api_key"
python register_postcube.py
```

The script creates:
- app record (`Postcube`)
- OAuth client for callback `http://localhost:8113/api/auth/callback`

It prints generated `BASALT_CLIENT_ID` and `BASALT_CLIENT_SECRET` for `backend/.env`.

## Git Ready

This project is ready for version control with:

- Root `.gitignore` configured for Go/Node/Python artifacts and local env files
- Project documentation in this README
- Basalt metadata included under `.basalt/`

Suggested first commit flow:

```powershell
git init
git add .
git commit -m "chore: initialize Postcube project"
```
