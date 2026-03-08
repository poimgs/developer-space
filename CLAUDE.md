# CLAUDE.md

## Project Structure

- **Frontend**: React + TypeScript + Vite app in `frontend/`
- **Backend**: Go API server in `cmd/`, `internal/`
- **Deployment**: Docker Compose with `Dockerfile.frontend` and `Dockerfile.api`

## Validation

Always run these checks before considering a change complete:

### Frontend
```bash
cd frontend && npx tsc -b          # TypeScript must compile with zero errors
cd frontend && npm test             # All Vitest tests must pass
cd frontend && npm run build        # Vite production build must succeed
```

### Backend
```bash
go build ./...                      # Go build must succeed
go test ./...                       # All Go tests must pass
```
