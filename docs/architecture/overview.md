# Architecture Documentation

## Technology Choices

### Web Framework: chi (Go)
- **Why**: Lightweight, idiomatic net/http compatible, middleware chaining, zero reflection
- **Alternative considered**: Gin (more magic, less control), Echo (heavier)
- **Key property**: Compatible with standard `net/http` Handler/Middleware signatures

### Database Access: sqlx + MySQL 8.0
- **Why**: Direct SQL control (no ORM magic), named parameters, struct scanning
- **Alternative considered**: GORM (implicit behavior, migration pain), raw database/sql (too verbose)
- **Key property**: Every query is visible and tunable

### WebSocket: gorilla/websocket
- **Why**: Battle-tested, clean API, WebSocket subprotocol support
- **Alternative considered**: nhooyr.io/websocket (newer, less ecosystem)

### Auth: JWT (HS256) + bcrypt
- **Why**: Stateless tokens suitable for WebSocket auth (token in connection query param)
- **Key property**: Server does NOT store sessions — token expiry is the only invalidation

### Real-time Voice/Video: LiveKit
- **Why**: Open-source WebRTC SFU, self-hosted with Docker Compose, JS/Go SDKs
- **Key property**: Game server only issues tokens — media flows peer-to-peer via LiveKit

## Communication Architecture

```
Client (Browser)                Server (Go)
     │                              │
     │──── REST ────►──────────► auth/login, register │
     │◄─── JWT token ◄───────────────│
     │                              │
     │──── WebSocket ───►──────────► join_room, room_action, chat │
     │◄─── state_update ◄────────────│
     │                              │
     │──── LiveKit ───►──────────► voice/video (via LiveKit SFU) │
```

- **REST**: Login, register, health check, admin CRUD
- **WebSocket**: Game state, room events, chat, real-time notifications
- **LiveKit**: Voice/video (separate channel, not through game server)

## Game Engine Design

```
GameEngine (interface)
  └── DoudizhuEngine (implementation)
      ├── deck.go — 54-card deck, shuffle, deal
      ├── hand.go — hand type evaluation
      ├── play.go — play validation (single, pair, triple, straight, bomb, etc.)
      └── engine.go — game loop: deal → bid → play → score
```

All game engines implement `GameEngine` interface in `internal/game/`, allowing future games (e.g., Upgrade, Bull Bull) by adding new implementations.

## Frontend State Flow

```
api-client (REST + JWT) ←→ localStorage (token)
     │
     ▼
ws-game (WebSocket client) ←→ Server Game Hub
     │
     ├──► components/game/ (card rendering, play area)
     ├──► components/chat/ (chat panel)
     └──► livekit-client (voice/video)
```

## Merging Strategy

- **Prefer squash merges** into main for feature branches
- **Commit style**: Conventional commits (`feat:`, `fix:`, `chore:`, `refactor:`)
- **Branch naming**: `feat/<feature-name>`, `fix/<bug-description>`, `chore/<task>`
