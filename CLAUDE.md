# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Multiplayer **Dou Di Zhu (斗地主)** card game platform with AI opponents, real-time WebSocket gameplay, and LiveKit voice/video. Starbucks-inspired design system. Admin panel for user/room/AI management.

## Architecture

**Server** (`server/`) — Go 1.26, chi router, gorilla/websocket, MySQL 8.0 via sqlx, JWT auth, bcrypt passwords

**Frontend** (`frontend/`) — Next.js 16, React 19, Tailwind CSS v4, TypeScript

### Server Package Layout

| Package | Purpose |
|---|---|
| `cmd/server/` | Entry point — wires config → DB → stores → hub → handlers → HTTP server |
| `internal/config/` | Env-based config via godotenv |
| `internal/model/` | MySQL stores (UserStore, GameStore, AIStore) + DB connection |
| `internal/auth/` | JWT token generation/validation + bcrypt password hashing |
| `internal/api/` | HTTP handlers — auth (register/login/guest), health, LiveKit token |
| `internal/api/middleware/` | Logging, CORS, JWT auth middleware, rate limiter |
| `internal/api/ws/` | WebSocket hub — room-based channels, read/write pumps, game message routing |
| `internal/api/admin/` | Admin REST endpoints — users, rooms, AI characters, LLM configs, scores |
| `internal/game/` | `GameEngine` interface + `RoomManager` for game lifecycle |
| `internal/game/doudizhu/` | Dou Di Zhu engine — cards (54-card deck), hand evaluation, bidding, play validation |
| `internal/ai/` | AI player with LLM provider abstraction (OpenAI-compatible) |

### Frontend Structure

| Path | Purpose |
|---|---|
| `app/auth/login/` + `register/` | Auth pages |
| `app/(main)/lobby/` | Game lobby |
| `app/(main)/room/[id]/` | Game room |
| `app/admin/` | Admin pages (layout with sidebar) — dashboard, users, rooms, ai-characters, llm-config, stats, scores |
| `components/game/` | Card rendering, hand display, game table, play area, action bar, player info |
| `components/chat/` | Chat panel, emoji picker, voice button |
| `components/ui/` | Reusable Button, Card, Input, AdminSidebar |
| `lib/api-client.ts` | REST API client (register, login) with JWT token management |
| `lib/ws-game.ts` | WebSocket game client — room join/leave, actions, chat, auto-reconnect |
| `lib/livekit-client.ts` | WebRTC voice/video via LiveKit |

### Game Flow

1. User logs in → gets JWT token
2. Joins/creates room via WebSocket (`join_room`)
3. Room auto-fills empty seats with AI bots
4. When 3 players ready → game starts (`game_start`)
5. Players send actions (`room_action`), server broadcasts `state_update`
6. Round end → scores calculated, room resets to `waiting`

### Key Infrastructure

- **Docker Compose**: MySQL 8.0 + LiveKit server + Go server
- **Migrations**: `server/migrations/` — `001_users`, `002_game_tables`, `003_ai_tables`, `004_nickname_unique`
- **LiveKit**: WebRTC SFU server for voice/video (`livekit.yaml`)
- **Card encoding**: IDs 0–51 standard cards (suit×13 + rank), 52=small joker, 53=big joker. Rank: 3→3, 4→4, ..., A→14, 2→15, small joker→16, big joker→17.

## Commands

### Server

```bash
cd server

# Build for current platform
go build -o server ./cmd/server/

# Build cross-platform
./build.sh linux amd64

# Run locally (requires MySQL on localhost:3306)
go run ./cmd/server/

# Run with docker-compose (MySQL + LiveKit + server)
docker compose up --build

# Run tests in a specific package
go test ./internal/game/doudizhu/...
go test ./internal/api/...
go test ./internal/model/...

# Run a single test
go test ./internal/game/doudizhu/ -run TestHandType

# Run with verbose output
go test -v ./internal/api/...

# Manual server start
./run.sh
```

### Frontend

```bash
cd frontend

# Development server (default :3000)
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint
npm run lint

# Type check
npx tsc --noEmit

# Run tests
npm test
```

### Database

Migrations run automatically via Docker Compose volume mount into `mysql` container. For local dev, apply manually:

```bash
mysql -u root -p pokeface < server/migrations/001_create_users.sql
```

### Environment

See `.env.example` for required variables:
- `DATABASE_DSN` — MySQL connection string
- `JWT_SECRET` — JWT signing key
- `ALLOWED_ORIGINS` — CORS origins (comma-separated)
- `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` / `LIVEKIT_HOST` — LiveKit credentials
- Frontend uses `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080`) and `NEXT_PUBLIC_WS_URL`

---

## Workflow Rules (MANDATORY)

1. **Plan first, execute second** — Before touching any code, state the plan (files to modify, approach). Wait for confirmation.
2. **One task at a time** — Do not proceed to the next task until the current one is verified complete.
3. **Report after each step** — After each logical step, report what was done and the result (success / failure / diff).
4. **No scope creep** — Stick to the stated task. If you discover related issues, flag them but do not fix unless explicitly instructed.
5. **Ask when unsure** — If a requirement is ambiguous, ask before proceeding.

## Coding Rules

- **No `any` types** — Use proper TypeScript types / Go interfaces everywhere.
- **No mass refactoring** — Refactor only the code directly relevant to the task at hand. Do not restructure unrelated code.
- **Server language** — Go only. No other languages in `server/`.
- **Frontend language** — TypeScript + Tailwind CSS v4 only. No plain JavaScript in `frontend/`.
- **No new dependencies** — Do not add new npm / Go modules without explicit approval.
- **Keep it simple** — Prefer straightforward solutions over clever abstractions.

## Verification (MANDATORY)

After any code change, run ALL applicable checks and show their output verbatim:

- **Server changes**: `go vet ./...` + `go test ./...` (or the relevant package(s))
- **Frontend changes**: `npm run lint` + `npx tsc --noEmit` + `npm test`
- **Both**: If both sides changed, run both sets of checks.

Do not claim success — show the actual output. If checks fail, fix before proceeding.

## Scope Constraints

**Allowed to modify:**
- `server/internal/` — any Go source (handlers, middleware, stores, game logic, AI)
- `frontend/` — any TypeScript/TSX/Tailwind source (pages, components, lib)
- Migrations that are already applied (only additive changes)
- `CLAUDE.md` itself

**Forbidden to modify without explicit instruction:**
- `server/migrations/` — creating new migration files (requires separate review)
- `server/cmd/server/` — main entry point changes
- `docker-compose.yml` or Dockerfile
- `livekit.yaml` or any LiveKit configuration
- CI/CD configuration (GitHub Actions, etc.)
- `server/go.mod` / `frontend/package.json` — dependency changes
- `.env.example` or environment variable schema changes
- Any build/deploy scripts (`build.sh`, `run.sh`)

## Task Granularity

- **One logical change per task** — A "task" is a single feature, bug fix, or refactoring.
- **Max 3 files per step** — If a change touches more than 3 files, split it into sub-steps.
- **Conventional commits** — Commit messages follow conventional commits format: `type(scope): description` (e.g. `feat(auth): add guest login`, `fix(ws): handle reconnect race`, `refactor(game): extract hand evaluator`).
